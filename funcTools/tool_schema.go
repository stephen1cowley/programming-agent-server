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

// Tool definition for editting the App.js file.
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
