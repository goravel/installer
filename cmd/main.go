package main

import (
	"fmt"
	"os"
	"os/exec"
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
	fmt.Println(ui.DefaultMessage.Render("This clones the repo and so on inside ", projectName))
	os := runtime.GOOS
	switch os {
	case "darwin":
		generateForUnix(projectName)
		return
	case "windows":
		generateForWindows(projectName)
		return
	default:
		fmt.Println(ui.ErrorMessage.Render("Unsupported platform. Closing"))
	}
}

func generateForUnix(projectName string) {
	clone := exec.Command("git", "clone", "https://github.com/goravel/goravel.git", projectName)

	fmt.Println(ui.DefaultMessage.Render("Cloning repo", projectName))
	err := clone.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println(ui.DefaultMessage.Render("Repo cloned"))

	remove_git := exec.Command("rm", "-rf", projectName+"/.git", projectName+"/.github")

	err = remove_git.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println(ui.DefaultMessage.Render("Git removal done"))
	fmt.Println(ui.DefaultMessage.Render("Instaling the Goravel"))
	install := exec.Command("go", "mod", "tidy")
	install.Dir = ("./" + projectName)
	install.Run()

	fmt.Print(ui.SuccessMessage.Render("Goravel installed sucessfuly !"))

	cp_env := exec.Command("cp", ".env.example", ".env")
	cp_env.Dir = ("./" + projectName)
	err = cp_env.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println(ui.DefaultMessage.Render("Generating app key "))

	app_key := exec.Command("go", "run", ".", "artisan", "key:", "generate")
	app_key.Dir = ("./" + projectName)
	err = app_key.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println("You can cd into your project and start developing ")
}

func generateForWindows(projectName string) {
	clone := exec.Command("git", "clone", "https://github.com/goravel/goravel.git", projectName)

	fmt.Println(ui.DefaultMessage.Render("Cloning repo", projectName))
	err := clone.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println(ui.DefaultMessage.Render("Repo cloned"))

	remove_git := exec.Command("Remove-Item", "-Path", "./"+projectName+"/.git", "./"+projectName+"/.github", "-Recursive", "-Force")

	err = remove_git.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println(ui.DefaultMessage.Render("Git removal done"))

	fmt.Println(ui.DefaultMessage.Render("Instaling the Goravel"))
	install := exec.Command("go", "mod", "tidy")
	install.Dir = ("./" + projectName)
	err = install.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println(ui.DefaultMessage.Render("Generating app key "))
	cp_env := exec.Command("cp", ".env.example", ".env")
	cp_env.Dir = ("./" + projectName)
	err = cp_env.Run()
	if err != nil {
		panic(err)
	}

	fmt.Print(ui.SuccessMessage.Render("Goravel installed sucessfuly !"))
	fmt.Println(ui.DefaultMessage.Render("Generating app key "))

	app_key := exec.Command("go", "run", ".", "artisan", "key:", "generate")
	app_key.Dir = ("./" + projectName)
	err = app_key.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println("You can cd into your project and start developing ")
}
