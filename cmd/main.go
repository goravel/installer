package main

import (
	"fmt"
	"github.com/goravel/installer/ui"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"runtime"
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
	Long:  `This is the goravel installer, build with love and care`,
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
	os := runtime.GOOS
	switch os {
	case "darwin", "linux":
		generateForUnix(projectName)
		return
	case "windows":
		generateForWindows(projectName)
		return
	default:
		fmt.Println(ui.ErrorMessage.Render("Following platform " + os + " is not supported. Installer is closing..."))
	}
}

func generateForUnix(projectName string) {
	clone := exec.Command("git", "clone", "https://github.com/goravel/goravel.git", projectName)
	fmt.Println(ui.DefaultMessage.Render("Creating a 'goravel/goravel' project at ./", projectName))
	if err := clone.Run(); err != nil {
		fmt.Printf(ui.ErrorMessage.Render("Error while generating the project : %s"), err)
		panic(err)
	}
	fmt.Println(ui.DefaultMessage.Render("Created project in ./", projectName))

	removeFiles := exec.Command("rm", "-rf", projectName+"/.git", projectName+"/.github")
	if err := removeFiles.Run(); err != nil {
		log.Panicf(ui.ErrorMessage.Render("Error happend while removing the files : %s"), err)
		panic(err)
	}
	fmt.Println(ui.SuccessMessage.Render("Git cleanup done."))

	fmt.Println(ui.DefaultMessage.Render("Instaling goravel/goravel"))
	install := exec.Command("go", "mod", "tidy")
	install.Dir = ("./" + projectName)
	if err := install.Run(); err != nil {
		log.Panicf(ui.ErrorMessage.Render("Error while installing the dependecies : %s"), err)
		panic(err)
	}
	fmt.Println(ui.SuccessMessage.Render("Goravel installed sucessfully!"))

	fmt.Println(ui.DefaultMessage.Render("Generating .env file..."))
	copyEnv := exec.Command("cp", ".env.example", ".env")
	copyEnv.Dir = ("./" + projectName)
	if err := copyEnv.Run(); err != nil {
		log.Panicf(ui.ErrorMessage.Render("Error while generating the .env file : %s"), err)
		panic(err)
	}
	fmt.Println(ui.SuccessMessage.Render(".env file generated sucessfully!"))

	fmt.Println(ui.DefaultMessage.Render("Generating app key "))
	initAppKey := exec.Command("go", "run", ".", "artisan", "key:", "generate")
	initAppKey.Dir = ("./" + projectName)
	if err := initAppKey.Run(); err != nil {
		log.Panicf(ui.ErrorMessage.Render("Error while generating the app key : %s"), err)
		panic(err)
	}

	fmt.Println(ui.InputLabelMuteTextStyle.Render("You can cd into your project and start developing "))
	fmt.Println(ui.InputLabelMuteTextStyle.Render("cd ./", projectName, "&& go run ."))

}

func generateForWindows(projectName string) {
	clone := exec.Command("git", "clone", "https://github.com/goravel/goravel.git", projectName)
	fmt.Println(ui.DefaultMessage.Render("Creating a 'goravel/goravel' project at ./", projectName))
	if err := clone.Run(); err != nil {
		fmt.Printf(ui.ErrorMessage.Render("Error while generating the project : %s"), err)
		panic(err)
	}
	fmt.Println(ui.SuccessMessage.Render("Created project in ./", projectName))

	removeFiles := exec.Command("Remove-Item", "-Path", "./"+projectName+"/.git", "./"+projectName+"/.github", "-Recursive", "-Force")
	if err := removeFiles.Run(); err != nil {
		log.Panicf(ui.ErrorMessage.Render("Error happend while removing the files : %s"), err)
		panic(err)
	}
	fmt.Println(ui.SuccessMessage.Render("Git cleanup done."))

	fmt.Println(ui.DefaultMessage.Render("Installing goravel/goravel"))
	install := exec.Command("go", "mod", "tidy")
	install.Dir = ("./" + projectName)
	if err := install.Run(); err != nil {
		log.Panicf(ui.ErrorMessage.Render("Error happend while installing goravel : %s"), err)
		panic(err)
	}

	fmt.Println(ui.DefaultMessage.Render("Generating .env file"))
	copyEnv := exec.Command("cp", ".env.example", ".env")
	copyEnv.Dir = ("./" + projectName)
	if err := copyEnv.Run(); err != nil {
		log.Panicf(ui.ErrorMessage.Render("Error while generating .env file : %s"), err)
		panic(err)
	}
	fmt.Print(ui.SuccessMessage.Render("Goravel installed sucessfully !"))

	fmt.Println(ui.DefaultMessage.Render("Generating app key "))
	initAppKey := exec.Command("go", "run", ".", "artisan", "key:", "generate")
	initAppKey.Dir = ("./" + projectName)
	if err := initAppKey.Run(); err != nil {
		log.Panicf(ui.ErrorMessage.Render("Error while generating application key : %s"), err)
		panic(err)
	}

	fmt.Println(ui.InputLabelMuteTextStyle.Render("You can cd into your project and start developing "))
	fmt.Println(ui.InputLabelMuteTextStyle.Render("cd ./", projectName, "&& go run ."))
}
