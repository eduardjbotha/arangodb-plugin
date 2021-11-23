package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	columnize "github.com/ryanuber/columnize"
)

const (
	helpHeader = `Usage: dokku repo[:COMMAND]

Runs commands that interact with the app's repo

Additional commands:`

	helpContent = `
    arangodb-plugin:test, prints test message
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	app := flag.Arg(1)
	service := flag.Arg(2)
	containerName := "arangodb-" + app
	environmentVariable := "ARANGODB_URL"

	switch cmd {
	case "arangodb-plugin:help":
		usage()
	case "help":
		fmt.Print("help called manually")
		fmt.Print(helpContent)
	case "arangodb-plugin:create":
		fmt.Println("triggered arangodb-plugin from: commands")
		out, err := exec.Command("docker", "images", "|", "grep", "\"arangodb\"", "|", "awk", "'{print $1}'").Output()
		if err != nil {
			fmt.Errorf("Docker image for ArangoDB not found. Please execute dokku plugin:install <repo>")
			os.Exit(1)
		}
		fmt.Println("Output: " + string(out))
		fmt.Println("Service: " + service)
		fmt.Println("Container: " + containerName)
		fmt.Println("ENV: " + environmentVariable)

	case "arangodb-plugin:delete":

		fmt.Println("called delete")
	case "arangodb-plugin:test":
		fmt.Println("triggered arangodb-plugin from: commands")
	default:
		dokkuNotImplementExitCode, err := strconv.Atoi(os.Getenv("DOKKU_NOT_IMPLEMENTED_EXIT"))
		if err != nil {
			fmt.Println("failed to retrieve DOKKU_NOT_IMPLEMENTED_EXIT environment variable")
			dokkuNotImplementExitCode = 10
		}
		os.Exit(dokkuNotImplementExitCode)
	}
}

func usage() {
	config := columnize.DefaultConfig()
	config.Delim = ","
	config.Prefix = "\t"
	config.Empty = ""
	content := strings.Split(helpContent, "\n")[1:]
	fmt.Println(helpHeader)
	fmt.Println(columnize.Format(content, config))
}
