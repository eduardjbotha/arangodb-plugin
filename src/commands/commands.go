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
    arangodb-plugin:info <app>, prints the container information
    arangodb-plugin:create <app>, creates the application container with a volume
    arangodb-plugin:delete <app>, deletes the application container and volume
    arangodb-plugin:test, prints test message
`
)

func executeBashCommand(command string, errorMessage string, continueAnyway bool) string {
	fmt.Println(fmt.Sprintf("executing bash command: %s", command))
	result, err := exec.Command("bash", "-c", command).Output()

	fmt.Println(fmt.Sprintf("Result %s, Error: %s", string(result), err))
	if err != nil {
		fmt.Errorf("Error executing command '%s': %s", command, err.Error())
		fmt.Errorf(errorMessage)
		if !continueAnyway {
			os.Exit(1)
		}
	}
	return strings.TrimSpace(string(result))
}

func getContainerId(containerName string) string {
	fmt.Println("stopping container: " + containerName)
	cmd := fmt.Sprintf("docker ps | grep %s | awk '{print $1}'", containerName)
	res := executeBashCommand(cmd, "Could not get container id", false)
	fmt.Println(fmt.Sprintf("Get container Id: %s", res))
	return res
}

func stopContainer(containerName string, remove bool) {
	fmt.Println(fmt.Sprintf("stop container: %s", containerName))
	idStr := getContainerId(containerName)
	if idStr != "" {
		fmt.Println("stop container")
		executeBashCommand(fmt.Sprintf("docker stop %s > /dev/null", idStr), "Could not stop container", false)
		if remove {
			fmt.Println("remove container")
			executeBashCommand(fmt.Sprintf("docker rm %s > /dev/null", idStr), "Could not remove container", true)
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
	environmentVariable := "ARANGODB_PASSWORD"

	dokkuRoot := os.Getenv("DOKKU_ROOT")
	pluginName := "arangodb"
	hostDirectory := fmt.Sprintf("%s/%s/%s", dokkuRoot, app, pluginName)
	password := ""

	switch cmd {
	case "arangodb-plugin:help":
		usage()
	case "help":
		fmt.Print("help called manually")
		fmt.Print(helpContent)
	case "arangodb-plugin:create":
		fmt.Println("triggered arangodb-plugin from: commands")

		cmd := "docker images | grep arangodb | awk '{print $1}'"
		image := executeBashCommand(cmd, "Docker image for ArangoDB not found. Please execute dokku plugin:install <repo>", false)
		fmt.Println("Output: " + string(image))

		fmt.Println("stopping container")
		stopContainer(containerName, false)
		fmt.Println("if hostdir exists")

		if _, err := os.Stat(hostDirectory); os.IsNotExist(err) {
			fmt.Println("host dir doesn't exist")

			executeBashCommand(fmt.Sprintf("mkdir -p %s && chown -R dokku:dokku %s", hostDirectory, hostDirectory), "Could not create directory", false)
		}
		fmt.Println("volumen name")

		volume := hostDirectory + ":/var/lib/arangodb"

		fmt.Println("create container cmd")
		createContainerCmd := fmt.Sprintf("docker run -d --name %s -p 8529:8529 -v %s -e ARANGO_RANDOM_ROOT_PASSWORD=1 arangodb/arangodb", containerName, volume)

		fmt.Println("execute create container")
		executeBashCommand(createContainerCmd, "Docker container could not be created", false)

		passwordCmd := fmt.Sprintf("docker logs %s | grep PASSWORD | awk '{print $4}'", containerName)
		fmt.Println("execute get password")
		password = executeBashCommand(passwordCmd, "Error getting password", true)

		fmt.Println("execute set password")
		executeBashCommand(fmt.Sprintf("dokku config:set \"%s\" %s=%s", app, environmentVariable, password), "Could not set arango environment variable", false)
		fmt.Println("finished")

		fmt.Println("Service: " + service)
		fmt.Println("Container created: " + containerName)
		fmt.Println("ENV: " + environmentVariable)

	case "arangodb-plugin:delete":

		fmt.Println("stopping container: " + containerName)
		stopContainer(containerName, true)

		fmt.Println("if host dir exists")
		if _, err := os.Stat(hostDirectory); !os.IsNotExist(err) {
			fmt.Println("delete host dir")
			executeBashCommand(fmt.Sprintf("rm -rf %s", hostDirectory), "Could not delete host directory", false)
		}

		fmt.Println("remove dokku config")
		executeBashCommand(fmt.Sprintf("dokku config:unset \"%s\" %s", app, environmentVariable), "Could not remove dokku configuration", false)
		fmt.Println(fmt.Sprintf("Container deleted: %s", containerName))

	case "arangodb-plugin:info":
		id := getContainerId(containerName)

		cmd := fmt.Sprintf("docker inspect %s | grep IPAddress | cut -d '\"' -f 4", id)
		ip := executeBashCommand(cmd, fmt.Sprintf("Docker container could not be inspected"), false)

		msg := `

				Host: %s
				Private ports: 8529
		`

		fmt.Println(fmt.Sprintf(msg, ip))
	case "arangodb-plugin:link":
		cmd := fmt.Sprintf("dokku link:create %s %s %s", app, containerName, pluginName)
		executeBashCommand(cmd, "Could not link", true)

		executeBashCommand(fmt.Sprintf("dokku config:set \"%s\" %s=%s", containerName, environmentVariable, password), "Could not set arango environment variable", false)
	case "arangodb-plugin:unlink":
		cmd := fmt.Sprintf("dokku link:delete %s %s %s", app, containerName, pluginName)
		executeBashCommand(cmd, "Could not unlink", true)
		executeBashCommand(fmt.Sprintf("dokku config:unset \"%s\" %s", containerName, environmentVariable), "Could not unset arango environment variable", false)
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
