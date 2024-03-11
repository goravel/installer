package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/goravel/installer/ui"
	"github.com/spf13/cobra"
)

const welcomeHeading = `
   ___   ___   ___    _ __   __ ___  _    
  / __| / _ \ | _ \  /_\\ \ / /| __|| |   
 | (_ || (_) ||   / / _ \\ V / | _| | |__ 
  \___| \___/ |_|_\/_/ \_\\_/  |___||____|
  `

var rootCmd = &cobra.Command{
	Use:   "goravel",
	Short: "Goravel installer",
	Long:  `This is the goravel installer , inspired by Laravel , build with love and care`,
}

var newProjectCmd = &cobra.Command{
	Use:   "new",
	Short: "Initiate a new goravel project",
	Long: `The new command is used to scaffold the base staring poin of you're 
  journey.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectName string // Default greeting
		fmt.Println(ui.LogoStyle.Render(welcomeHeading))
		if len(args) > 0 {
			projectName = args[0]
			generate(projectName)
		} else {
			// here ask for user input and generate the project
			fmt.Println(ui.InputLabelStyle.Render("What is the name of your project?"))
			fmt.Println(ui.InputLabelMuteTextStyle.Render("E.g my-new-app"))
			fmt.Print(ui.InputStyle.Render(">"))
			fmt.Scan(&projectName)
			generate(projectName)
		}
		cmd.Version = "0.0.1"

		return nil
	},
}

func main() {
	rootCmd.AddCommand(newProjectCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generate(projectName string) {
	fmt.Println(ui.DefaultMessage.Render("This clones the repo and so on inside ", projectName))
	os := runtime.GOOS
	switch os {
	case "darwin":
		fmt.Println(ui.DefaultMessage.Render("Generating ... .. . .. . .. proj for  ", os))
		return
	case "windows":
		fmt.Println(ui.DefaultMessage.Render("Generating ... .. . .. . .. proj for  ", os))
		return
	default:
		fmt.Println(ui.ErrorMessage.Render("Unsupported platform. Closing"))
	}
}
