package funcTools

import (
	"fmt"
	"os/exec"
)

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
