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
    arangodb-plugin:help, displays this help message
    arangodb-plugin:info, prints the container information
    arangodb-plugin:create <app>, creates the application container with a volume
    arangodb-plugin:delete <app>, deletes the application container and volume
    arangodb-plugin:test, prints test message
`
)

func executeBashCommand(command string, errorMessage string) string {
	result, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		fmt.Errorf("Error executing command '%s': %s", command, err)
		fmt.Errorf(errorMessage)
		os.Exit(1)
	}
	return string(result)
}

func getContainerId(containerName string) string {
	cmd := fmt.Sprintf("docker ps | grep %s | awk '{print $1}'", containerName)
	return executeBashCommand(cmd, "Could not get container id")
}

func stopContainer(containerName string, remove bool) {
	idStr := getContainerId(containerName)
	if idStr != "" {
		exec.Command(fmt.Sprintf("bash", "-c", "docker stop %s > /dev/null", idStr))
		if remove {
			exec.Command(fmt.Sprintf("bash", "-c", "docker rm %s > /dev/null", idStr))
		}
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	app := flag.Arg(1)
	service := flag.Arg(2)
	containerName := "arangodb-" + app
	environmentVariable := "ARANGODB_URL"

	dokkuRoot := os.Getenv("DOKKU_ROOT")
	pluginName := "arangodb"
	hostDirectory := fmt.Sprintf("%s/%s/%s", dokkuRoot, app, pluginName)

	switch cmd {
	case "arangodb-plugin:help":
		usage()
	case "help":
		fmt.Print("help called manually")
		fmt.Print(helpContent)
	case "arangodb-plugin:create":
		fmt.Println("triggered arangodb-plugin from: commands")

		cmd := "docker images | grep arangodb | awk '{print $1}'"
		image, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			fmt.Errorf("Docker image for ArangoDB not found. Please execute dokku plugin:install <repo>")
			os.Exit(1)
		}
		fmt.Println("Output: " + string(image))

		stopContainer(containerName, false)

		if _, err := os.Stat(hostDirectory); os.IsNotExist(err) {
			executeBashCommand(fmt.Sprintf("mkdir -p %s && chown -R dokku:dokku %s", hostDirectory, hostDirectory), "Could not create directory")
		}
		volume := hostDirectory + ":/var/lib/arangodb"

		createContainerCmd := fmt.Sprintf("docker run -d --name %s -p 8529:8529 -v %s -e ARANGO_RANDOM_ROOT_PASSWORD=1 arangodb/arangodb", containerName, volume)

		_, err = exec.Command("bash", "-c", createContainerCmd).Output()
		if err != nil {
			fmt.Errorf("Docker container could not be created: %s", err)
			os.Exit(1)
		}

		fmt.Println("Service: " + service)
		fmt.Println("Container created: " + containerName)
		fmt.Println("ENV: " + environmentVariable)

	case "arangodb-plugin:delete":

		stopContainer(containerName, false)

		if _, err := os.Stat(hostDirectory); !os.IsNotExist(err) {
			executeBashCommand(fmt.Sprintf("rm -rf %s", hostDirectory), "Could not delete host directory")
		}

		executeBashCommand(fmt.Sprintf("dokku config:unset \"%s\", %s", app, environmentVariable), "Could not remove dokku configuration")
		fmt.Sprintf("Container deleted: %s", containerName)

	case "arangodb-plugin:info":
		id := getContainerId(containerName)

		cmd := fmt.Sprintf("docker inspect %s | grep IPAddress | cut -d '\"' -f 4", id)
		ip := executeBashCommand(cmd, fmt.Sprintf("Docker container could not be inspected"))

		msg := `

				Host: %s
				Private ports: 8529
		`

		fmt.Sprintf(msg, ip)
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
