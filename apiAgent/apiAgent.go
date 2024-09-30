package apiAgent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sashabaranov/go-openai"
	awsHandlers "github.com/stephen1cowley/programming-agent-server/awsHandlers"
	funcTools "github.com/stephen1cowley/programming-agent-server/funcTools"
)

// msgsSchema is a single message, with role typically being "user" or "ai"
type msgSchema struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

// msgsSchema is a list of user and ai messages.
type msgsSchema struct {
	Messages []msgSchema `json:"messages"`
}

// secretSchema is the schema returned by the AWS Secrets Manager
type secretSchema struct {
	OpenAIAPI string `json:"OpenAIAPI"`
}

// deleteFileSchema is the schema of the incoming POST request to delete a file from S3.
type deleteFileSchema struct {
	FileName string `json:"fileName"`
}

// Global variables
const TEST_USER_ID = "123"
const ECS_CLUSTER_NAME = "ProjectCluster2"

var apiKey string
var client openai.Client

var myTools []openai.Tool
var secretData secretSchema

// APiAgent sets up the Programming Agent http server on port 80
func ApiAgent() {
	err := onRestart()
	if err != nil {
		log.Fatalf("Error restarting the application, %v", err)
	}

	http.Handle("/api/test/", http.HandlerFunc(apiTestHandler))
	http.Handle("/api/message", corsMiddleware(http.HandlerFunc(apiMessageHandler)))
	http.Handle("/api/restart", corsMiddleware(http.HandlerFunc(apiRestartHandler)))
	http.Handle("/api/upload", corsMiddleware(http.HandlerFunc(apiUploadHandler)))
	http.Handle("/api/imdel", corsMiddleware(http.HandlerFunc(apiImdelHandler)))
	http.Handle("/api/reset", corsMiddleware(http.HandlerFunc(apiResetHandler)))

	log.Println("Server listening on :80")
	err = http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatalf("Error starting server on :80: %v\n", err)
	}
}

// onRestart initializes a variety of different variables and AWS settings, upon reloading the frontend
func onRestart() error {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
		return err
	}

	awsHandlers.InitDynamo(cfg)

	// Concurrently create an S3 client and DynamoDB client
	go awsHandlers.InitS3(cfg)

	// Create a Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)

	// Create the input for the GetSecretValue API call
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("open-ai-api-key"),
	}

	// Retrieve the secret value
	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Printf("failed to retrieve secret, %v", err)
		return err
	}

	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		log.Printf("Failed to unmarshal secret: %v", err)
		return err
	}

	// Initialise chat variables
	// Starting system message always prepended to list of messages
	apiKey = secretData.OpenAIAPI
	client = *openai.NewClient(apiKey)

	myTools = []openai.Tool{funcTools.AppJSEdit, funcTools.AppCSSEdit}

	// No errors on startup
	return nil
}

// apiResetHandler initializes a variety of different variables and AWS settings, upon pressing the reset button on the frontend
func apiResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		currUserID := r.Header.Get("username")
		log.Println("RESET Request from user", currUserID)

		// Get the previous UserState
		currUserState, err := awsHandlers.DynamoGetUser(currUserID)
		if err != nil {
			http.Error(w, "Failed to find user of given credentials", http.StatusInternalServerError)
			log.Printf("Failed to find user of given credentials %v\n", err)
			return
		}

		// Reset everything other than the UserID and the Fargate task
		freshUserState := awsHandlers.UserState{}
		freshUserState.UserID = currUserState.UserID
		freshUserState.FargateTaskARN = currUserState.FargateTaskARN

		err = awsHandlers.DynamoPutUser(freshUserState)
		if err != nil {
			http.Error(w, "Failed to reset user info", http.StatusInternalServerError)
			log.Printf("Failed to reset user info %v", err)
		}

		err = awsHandlers.DeleteAllFromS3(currUserID)
		if err != nil {
			http.Error(w, "Failed to delete all images from S3", http.StatusInternalServerError)
			log.Printf("Failed to delete all items in the S3 folder: %v\n", err)
		}
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println(w, "Method not allowed")
	}
}

// apiMessageHandler formulates the query to ChatGPT and responds accordingly.
// It is executed whenever a user sends a message to the AI.
func apiMessageHandler(w http.ResponseWriter, r *http.Request) {
	var editAppJSResp funcTools.ArgsAppJS
	var editAppCSSResp funcTools.ArgsAppCSS

	var currUserState *awsHandlers.UserState
	var err error

	if r.Method == http.MethodPost {
		currUserID := r.Header.Get("username")
		log.Println("Request from user", currUserID)

		var startSysMsg = openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleSystem,
			Content: `You are a helpful software engineer.
			Currently we are working on a fresh React App boilerplate, with access to react-bootstrap, bootstrap and react-router-dom modules.
			You are able to change App.js and App.css, for a web app which is updated live.
		
			The user knows nothing about computer programming.
			Therefore you must not say what you are doing under the hood when it comes to updating code.
			You must be helpful and polite, and always give a brief description of what the website you created should look like.
			But remember, don't mention App.js or App.css or what you've done to the code, as this means nothing to the user!
	
			You also have access to an S3 bucket folder for images https://my-programming-agent-img-store.s3.eu-west-2.amazonaws.com/uploads/
			
			` + currUserID + "/.",
		}

		// Get the previous UserState
		currUserState, err = awsHandlers.DynamoGetUser(currUserID)
		if err != nil {
			http.Error(w, "Failed to find user of given credentials", http.StatusInternalServerError)
			log.Printf("Failed to find user of given credentials %v\n", err)
			return
		}

		currUserState.DirectoryState.S3Images, err = awsHandlers.ListAllInS3("uploads/" + currUserState.UserID)
		if err != nil {
			http.Error(w, "Error finding images", http.StatusInternalServerError)
			log.Printf("Error finding images %v\n", err)
			return
		}
		fmt.Println("Current images are", currUserState.DirectoryState.S3Images)

		var requestData msgsSchema
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Error decoding request data", http.StatusBadRequest)
			fmt.Fprintln(w, "Error decoding request data:", err)
			return
		}

		// Process data after successful unmarshalling
		text := requestData.Messages[0].Text
		currUserState.Messages = append(currUserState.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		// System message at the end describing the current state of the files
		endSysMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: currUserState.DirectoryState.CreateSysMsgState(),
		}

		// Start and end system message with user/machine communication sandwiched inbetween
		messagesWithSys := append(append([]openai.ChatCompletionMessage{startSysMsg}, currUserState.Messages...), endSysMsg)
		fmt.Println(messagesWithSys)

		// Define a regular expression pattern to match everything between backticks
		re := regexp.MustCompile("```[^```]+```")

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		resp, err := client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:       openai.GPT4o,
				Messages:    messagesWithSys,
				Tools:       myTools,
				Temperature: 0.8,
				// ToolChoice: "required",
			},
		)
		cancel()

		if err != nil {
			http.Error(w, "ChatCompletion error", http.StatusInternalServerError)
			log.Printf("ChatCompletion error: %v\n", err)
			return
		}

		content := resp.Choices[0].Message.Content

		// Hard code a response if only tool calls were made
		if content == "" {
			content = "Done!"
		}

		tool_calls := resp.Choices[0].Message.ToolCalls

		fmt.Println(content)

		if len(tool_calls) != 0 {
			fmt.Println("Now making any tool calls ...")
		}

		// Replace all occurrences of stuff between ```...```
		content = re.ReplaceAllString(content, "")

		for _, val := range tool_calls {
			switch val.Function.Name {
			case "app_js_edit_func":
				fmt.Println("Updating App.js ...")
				json.Unmarshal([]byte(val.Function.Arguments), &editAppJSResp)
				awsHandlers.EditAppJS(
					editAppJSResp.AppJSCode,
					currUserID,
				)
				currUserState.DirectoryState.AppJSCode = editAppJSResp.AppJSCode
			case "app_css_edit_func":
				fmt.Println("Updating App.css ...")
				json.Unmarshal([]byte(val.Function.Arguments), &editAppCSSResp)
				awsHandlers.EditAppCSS(
					editAppCSSResp.AppCSSCode,
					currUserID,
				)
				currUserState.DirectoryState.AppCSSCode = editAppCSSResp.AppCSSCode
			}
		}

		if len(currUserState.Messages) >= 10 {
			currUserState.Messages = currUserState.Messages[2:]
		}

		currUserState.Messages = append(currUserState.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})

		w.Header().Set("Content-Type", "application/json")

		// Update the UserState now that messages have been added and file contents changed
		err = awsHandlers.DynamoPutUser(*currUserState)
		if err != nil {
			log.Printf("Failed to add fresh user %v", err)
		}

		// Create output and respond (same as input schema for now...)
		jsonResponse := msgSchema{Role: "ai", Text: content}
		if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
			http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
			log.Println(w, "Error encoding JSON response", err)
			return
		}
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println(w, "Method not allowed")
	}
}

// apiRestartHandler handles requests that occur each time the frontend is reloaded
func apiRestartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		err := onRestart()
		if err != nil {
			http.Error(w, "Error restarting the application", http.StatusInternalServerError)
			log.Fatalf("Error restarting the application, %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println(w, "Method not allowed")
	}
}

// apiImdelHandler handles requests to delete an image from S3.
func apiImdelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		currUserID := r.Header.Get("username")
		log.Println("Request from user", currUserID)

		var deleteRequest deleteFileSchema
		err := json.NewDecoder(r.Body).Decode(&deleteRequest)
		if err != nil {
			http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
			fmt.Fprintln(w, "Error decoding request data:", err)
			return
		}
		fileToDelete := deleteRequest.FileName
		err = awsHandlers.DeleteFromS3(fileToDelete, currUserID)
		if err != nil {
			http.Error(w, "Error deleting file", http.StatusInternalServerError)
			log.Println("Error deleting file, ", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println(w, "Method not allowed")
	}
}

// apiUploadHandler handles requests to upload an image from S3.
func apiUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		currUserID := r.Header.Get("username")
		log.Println("Request from user", currUserID)

		// Parse the form with a max size of 10MB
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			http.Error(w, "Unable to process file", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving the file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Upload to S3
		fileURL, err := awsHandlers.UploadToS3(file, handler, currUserID)
		if err != nil {
			http.Error(w, "Failed to upload file to S3", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("File uploaded successfully: %s", fileURL)))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println(w, "Method not allowed")
	}
}

// apiTestHandler handles health check requests of the ALB.
func apiTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		log.Println("Test request received...")
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println(w, "Method not allowed")
	}
}

// corsMiddleware adds necessary headers for CORS policy and preflight requests.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "https://stephencowley.com")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, username")

		// Handle preflight request (OPTIONS)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
