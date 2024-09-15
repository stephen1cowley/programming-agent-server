package cliAgent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sashabaranov/go-openai"
)

var editAppJSResp ArgsAppJS
var editAppCSSResp ArgsAppCSS
var newFileResp ArgsCreateFile
var libsResp ArgsLibraries

func CliAgent() {

	// Start off by cleaning the React App source code
	cmd := exec.Command("shell_script/onStartup.sh")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(output))

	currDirState := DirectoryState{
		AppJSCode:  "",
		AppCSSCode: "",
	}

	apiKey := os.Getenv("OPEN_AI_API_KEY")
	client := openai.NewClient(apiKey)
	messages := make([]openai.ChatCompletionMessage, 0)

	// Starting system message always prepended to list of messages
	startSysMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful software engineer. Currently we are working on a fresh React App boilerplate, with access to Bootstrap 5 styles. You are able to change App.js and App.css. You are able to create new JavaScript files to assist you in creating the application, ensure these are correctly imported into App.js.",
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")
	myTools := []openai.Tool{AppJSEdit, AppCSSEdit, NewJsonFile}

	// Define a regular expression pattern to match everything between backticks
	re := regexp.MustCompile("```[^```]+```")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
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
			continue
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
				EditAppJS(
					editAppJSResp.AppJSCode,
				)
				currDirState.AppJSCode = editAppJSResp.AppJSCode
			case "app_css_edit_func":
				fmt.Println("Updating App.css ...")
				json.Unmarshal([]byte(val.Function.Arguments), &editAppCSSResp)
				EditAppCSS(
					editAppCSSResp.AppCSSCode,
				)
				currDirState.AppCSSCode = editAppCSSResp.AppCSSCode
			case "new_js_file_func":
				fmt.Println("Creating new JS file ...")
				json.Unmarshal([]byte(val.Function.Arguments), &newFileResp)
				CreateJSFile(
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
					FileState{
						FileName: newFileResp.FileName,
						FileCode: newFileResp.FileContent,
					})
			case "libraries_func":
				fmt.Println("Importing libraries ...")
				json.Unmarshal([]byte(val.Function.Arguments), &libsResp)
				InstallLibraries(
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
	}
}
