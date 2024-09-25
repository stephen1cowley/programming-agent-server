package funcTools

// Filestate holds a file's name and the code within it
type FileState struct {
	FileName string
	FileCode string
}

// DirectoryState holds the state of current JS, CSS files and the file names of available images
type DirectoryState struct {
	AppJSCode  string
	AppCSSCode string
	OtherFiles []FileState
	S3Images   []string
}

// CreateSysMsgState formulates a user message to be sent each time to the LLM
func (cd DirectoryState) CreateSysMsgState() (sysMsg string) {
	if len(cd.S3Images) == 0 {
		sysMsg += "Currently, there are NO images in the S3 folder."
	}
	for _, imName := range cd.S3Images {
		sysMsg += "Currently, the images available to you in the S3 folder are:"
		sysMsg += "\n" + imName
	}
	sysMsg += "\n\n"
	sysMsg += "The current file contents are as follows:\n\n"
	sysMsg += "`App.js`:\n\n```jsx\n"
	sysMsg += cd.AppJSCode
	sysMsg += "\n```css\n\n`App.css`:\n\n```\n"
	sysMsg += cd.AppCSSCode
	sysMsg += "\n```"
	for _, file := range cd.OtherFiles {
		sysMsg += "\n\n`" + file.FileName + ".js`:\n\n```jsx\n" + file.FileCode + "\n```"
	}
	return
}
