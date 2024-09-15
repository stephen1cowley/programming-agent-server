package funcTools

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type ArgsAppJS struct {
	AppJSCode string `json:"appjscode"`
}

type ArgsAppCSS struct {
	AppCSSCode string `json:"appcsscode"`
}

type ArgsCreateFile struct {
	FileName    string `json:"filename"`
	FileContent string `json:"filecontent"`
}

type ArgsLibraries struct {
	Libraries string `json:"libraries"`
}

// Tool definition for editting the App.js file
var (
	AppJSjsonSchema = jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"appjscode": {
				Type:        jsonschema.String,
				Description: "The new App.js React code of the website. You MUST define a function App() and export default App in this file.",
			},
		},
	}

	AppJSEditFuncDef = openai.FunctionDefinition{
		Name:        "app_js_edit_func",
		Description: "Replaces the App.js code of the React website with the inputted code.",
		Parameters:  &AppJSjsonSchema,
	}

	AppJSEdit = openai.Tool{
		Type:     openai.ToolType("function"),
		Function: &AppJSEditFuncDef,
	}
)

// Tool definition for editting the App.css file
var (
	AppCSSjsonSchema = jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"appcsscode": {
				Type:        jsonschema.String,
				Description: "The new App.css React code of the website.",
			},
		},
	}

	AppCSSEditFuncDef = openai.FunctionDefinition{
		Name:        "app_css_edit_func",
		Description: "Replaces the App.css code of the React website with the inputted code.",
		Parameters:  &AppCSSjsonSchema,
	}

	AppCSSEdit = openai.Tool{
		Type:     openai.ToolType("function"),
		Function: &AppCSSEditFuncDef,
	}
)

// Tool definition for creating another file
var (
	NewFileJsonSchema = jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"filename": {
				Type:        jsonschema.String,
				Description: "The name of the new JavaScript file to create for the app, without the '.js' at the end, e.g. 'utils'",
			},
			"filecontent": {
				Type:        jsonschema.String,
				Description: "The content of the new JavaScript file.",
			},
		},
	}

	NewFileFuncDef = openai.FunctionDefinition{
		Name:        "new_js_file_func",
		Description: "Creates a new JavaScript file of a given name and content to assist in creating functionality for the React app.",
		Parameters:  &NewFileJsonSchema,
	}

	NewJsonFile = openai.Tool{
		Type:     openai.ToolType("function"),
		Function: &NewFileFuncDef,
	}
)

// Tool definition for importing libraries
var (
	LibrariesJsonSchema = jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"libraries": {
				Type:        jsonschema.String,
				Description: "String of a library that require importing",
			},
		},
	}

	LibrariesFuncDef = openai.FunctionDefinition{
		Name:        "libraries_func",
		Description: "Imports the required libraries into the Node React project usign npm install ...",
		Parameters:  &LibrariesJsonSchema,
	}

	ImportLibraries = openai.Tool{
		Type:     openai.ToolType("function"),
		Function: &LibrariesFuncDef,
	}
)
