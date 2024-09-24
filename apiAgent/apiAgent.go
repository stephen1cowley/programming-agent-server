package apiAgent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sashabaranov/go-openai"
	funcTools "github.com/stephen1cowley/programming-agent-server/funcTools"
	s3handler "github.com/stephen1cowley/programming-agent-server/s3Handler"
)

type msgSchema struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type msgsSchema struct {
	Messages []msgSchema `json:"messages"`
}

type secretSchema struct {
	OpenAIAPI string `json:"OpenAIAPI"`
}

type deleteFileSchema struct {
	FileName string `json:"fileName"`
}

// Global variables (for now...)
var apiKey string
var client openai.Client
var messages []openai.ChatCompletionMessage
var currDirState funcTools.DirectoryState
var startSysMsg openai.ChatCompletionMessage
var myTools []openai.Tool
var editAppJSResp funcTools.ArgsAppJS
var editAppCSSResp funcTools.ArgsAppCSS
var newFileResp funcTools.ArgsCreateFile
var libsResp funcTools.ArgsLibraries
var secretData secretSchema

func ApiAgent() {
	onRestart()

	// Create a new router
	// router := http.NewServeMux()

	// Apply CORS middleware to all routes
	http.HandleFunc("/api/test/", apiTestHandler)

	// Now we can begin the conversation by opening up the server!
	http.HandleFunc("/api/message", apiMessageHandler)
	http.HandleFunc("/api/restart", apiRestartHandler)
	http.HandleFunc("/api/upload", apiUploadHandler)
	http.HandleFunc("/api/imdel", apiImdelHandler)

	fmt.Println("Server listening on :80")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func onRestart() {
	// Start off by cleaning the React App source code
	cmd := exec.Command("/home/ubuntu/shell_script/onStartup.sh")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(output))

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-2"))
	if err != nil {
		fmt.Printf("unable to load SDK config, %v", err)
	}

	// Create a Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)

	// Create the input for the GetSecretValue API call
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("open-ai-api-key"),
	}

	// Retrieve the secret value
	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		fmt.Printf("failed to retrieve secret, %v", err)
	}

	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		log.Fatalf("Failed to unmarshal secret: %v", err)
	}

	// Initialise chat variables
	apiKey = secretData.OpenAIAPI
	client = *openai.NewClient(apiKey)
	messages = make([]openai.ChatCompletionMessage, 0)
	currDirState = funcTools.DirectoryState{} // i.e. initially empty
	startSysMsg = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful software engineer. Currently we are working on a fresh React App boilerplate, with access to Bootstrap 5 styles. You are able to change App.js and App.css. You are able to create new JavaScript files to assist you in creating the application, ensure these are correctly imported into App.js. You also have access to an S3 bucket for images: https://my-programming-agent-img-store.s3.eu-west-2.amazonaws.com/.",
	} // Starting system message always prepended to list of messages
	myTools = []openai.Tool{funcTools.AppJSEdit, funcTools.AppCSSEdit, funcTools.NewJsonFile}
}

func apiMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://stephencowley.com") // Replace with your allowed origin(s)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodPost {

		var requestData msgsSchema
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Error decoding request data:", err)
			return
		}

		// Process data after successful unmarshalling
		text := requestData.Messages[0].Text
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		// System message at the end describing the current state of the files
		endSysMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: currDirState.CreateSysMsgState(),
		}

		// Start and end system message with user/machine communication sandwiched inbetween
		messagesWithSys := append(append([]openai.ChatCompletionMessage{startSysMsg}, messages...), endSysMsg)
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
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}

		content := resp.Choices[0].Message.Content

		tool_calls := resp.Choices[0].Message.ToolCalls

		fmt.Println(content)
		fmt.Println(messages)

		if len(tool_calls) != 0 {
			fmt.Println("Now making any tool calls ...")
		}

		// Replace all occurrences of stuff between ```...```
		content = re.ReplaceAllString(content, "")

		for _, val := range tool_calls {
		outerSwitch:
			switch val.Function.Name {
			case "app_js_edit_func":
				fmt.Println("Updating App.js ...")
				json.Unmarshal([]byte(val.Function.Arguments), &editAppJSResp)
				funcTools.EditAppJS(
					editAppJSResp.AppJSCode,
				)
				currDirState.AppJSCode = editAppJSResp.AppJSCode
			case "app_css_edit_func":
				fmt.Println("Updating App.css ...")
				json.Unmarshal([]byte(val.Function.Arguments), &editAppCSSResp)
				funcTools.EditAppCSS(
					editAppCSSResp.AppCSSCode,
				)
				currDirState.AppCSSCode = editAppCSSResp.AppCSSCode
			case "new_js_file_func":
				fmt.Println("Creating new JS file ...")
				json.Unmarshal([]byte(val.Function.Arguments), &newFileResp)
				funcTools.CreateJSFile(
					newFileResp,
				)
				for i, file := range currDirState.OtherFiles {
					if newFileResp.FileName == file.FileName {
						currDirState.OtherFiles[i].FileCode = newFileResp.FileContent
						break outerSwitch
					}
				}
				// File doesn't yet exist in the list; append this new file.
				currDirState.OtherFiles = append(
					currDirState.OtherFiles,
					funcTools.FileState{
						FileName: newFileResp.FileName,
						FileCode: newFileResp.FileContent,
					})
			case "libraries_func":
				fmt.Println("Importing libraries ...")
				json.Unmarshal([]byte(val.Function.Arguments), &libsResp)
				funcTools.InstallLibraries(
					libsResp,
				)
			}
		}

		if len(messages) >= 10 {
			messages = messages[2:]
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})

		fmt.Println(messages)

		// Create output and respond (same as input schema for now...)
		jsonResponse := msgSchema{Role: "ai", Text: content}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error encoding JSON response", err)
			return
		}

	} else if r.Method == http.MethodOptions {
		// Handle preflight request
		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed") // server debug message
	}
}

func apiRestartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://stephencowley.com") // Replace with your allowed origin(s)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	w.WriteHeader(http.StatusOK)

	if r.Method == http.MethodPut {
		onRestart()
	} else if r.Method == http.MethodOptions {
		return
	}
}

func apiImdelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://stephencowley.com") // Replace with your allowed origin(s)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	} else if r.Method == http.MethodPost {
		var deleteRequest deleteFileSchema
		err := json.NewDecoder(r.Body).Decode(&deleteRequest)
		if err != nil {
			http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
		}
		fileToDelete := deleteRequest.FileName
		err = s3handler.DeleteFromS3(fileToDelete)
		if err != nil {
			fmt.Println("Error deleting file, ", err)
			http.Error(w, "Error deleting file", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed") // server debug message
	}
}

func apiUploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://stephencowley.com") // Replace with your allowed origin(s)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "text/plain")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	} else if r.Method == http.MethodPost {
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
		fileURL, err := s3handler.UploadToS3(file, handler)
		if err != nil {
			http.Error(w, "Failed to upload file to S3", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(fmt.Sprintf("File uploaded successfully: %s", fileURL)))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed") // server debug message
	}
}

func apiTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Println("Test request received...")
		w.WriteHeader(http.StatusOK)
		return
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Replace with your allowed origin(s)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}
