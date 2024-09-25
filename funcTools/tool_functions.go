package funcTools

import (
	"fmt"
	"os/exec"
)

// EditAppJS runs shell script to edit App.js
func EditAppJS(AppJSCode string) {
	// Run the shell script with the variable value
	cmd := exec.Command("~/shell_script/editAppJS.sh", AppJSCode)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the output from the shell script
	fmt.Println(string(output))
}

// EditAppCSS runs shell script to edit App.css
func EditAppCSS(AppCSSCode string) {
	// Run the shell script with the variable value
	cmd := exec.Command("~/shell_script/editAppCSS.sh", AppCSSCode)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the output from the shell script
	fmt.Println(string(output))
}

// CreateJSFile runs shell script to create a new JS file
func CreateJSFile(CreateJSFileArgs ArgsCreateFile) {
	// Run the shell script with the variable value
	cmd := exec.Command("~/shell_script/createJSFile.sh", CreateJSFileArgs.FileName, CreateJSFileArgs.FileContent)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the output from the shell script
	fmt.Println(string(output))
}

// InstallLibraries installs requested libraries using shellscript
func InstallLibraries(IntallLibrariesArgs ArgsLibraries) {
	// libStr := ""
	// for _, lib := range IntallLibrariesArgs.Libraries {
	// 	libStr += " "
	// 	libStr += lib
	// }
	libStr := IntallLibrariesArgs.Libraries
	// Run the shell script to import the required libraries
	cmd := exec.Command("~/shell_script/importLibs.sh", libStr)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// Print the output from the shell script
	fmt.Println(string(output))
}
