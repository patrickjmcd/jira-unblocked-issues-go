package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/andygrunwald/go-jira"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

func getUnspecifiedKey(key string) string {
	var byteRead []byte
	var stringRead string
	var err error
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", key)
	if key == "Jira Password" {
		byteRead, err = terminal.ReadPassword(int(syscall.Stdin))
		stringRead = string(byteRead)
	} else {
		stringRead, err = reader.ReadString('\n')
	}

	if err != nil {
		log.Fatalf("You need to specify a %s\n", key)
	}
	trimmedVal := strings.TrimSuffix(stringRead, "\n")
	return trimmedVal
}

func getEnvVariablesOrAsk() (string, string, string) {
	var jiraURL string
	var jiraUsername string
	var jiraPassword string

	viper.SetEnvPrefix("jira")
	viper.BindEnv("username")
	viper.BindEnv("url")
	viper.BindEnv("password")

	jiraURL = viper.GetString("url")
	if !viper.IsSet("url") {
		jiraURL = getUnspecifiedKey("Jira URL")
		os.Setenv("JIRA_URL", jiraURL)
	}

	jiraUsername = viper.GetString("username")
	if !viper.IsSet("username") {
		jiraUsername = getUnspecifiedKey("Jira Username")
		os.Setenv("JIRA_USERNAME", jiraUsername)
	}

	jiraPassword = viper.GetString("password")
	if !viper.IsSet("password") {
		jiraPassword = getUnspecifiedKey("Jira Password")
		os.Setenv("JIRA_PASSWORD", jiraPassword)
	}

	return jiraURL, jiraUsername, jiraPassword
}

func main() {
	var projectName string
	var verbose bool

	flag.StringP("project", "P", "", "The Jira code for the project to monitor")
	flag.BoolP("verbose", "v", false, "Print all issues to the console")
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	projectName = viper.GetString("project")
	verbose = viper.GetBool("verbose")

	url, username, password := getEnvVariablesOrAsk()
	transport := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}
	jiraClient, err := jira.NewClient(transport.Client(), url)
	if err != nil {
		log.Fatal("Couldn't log on to the Jira server.")
	}

	// checkResolvedLinkedIssuesForProject(jiraClient, projectName, verbose)
	issuesWithResolved := getResolvedLinkedIssuesForProject(jiraClient, projectName, verbose)
	if len(issuesWithResolved) > 0 {
		color.Red("------------------------------------------------------")
		color.Red("   The following %d issues have completed linked issues  ", len(issuesWithResolved))
		color.Red("------------------------------------------------------")
		for _, issue := range issuesWithResolved {
			color.Red("[%s] %s - %s/browse/%s", issue.Key, issue.Fields.Summary, url, issue.Key)
		}
		color.Red("------------------------------------------------------")
	} else {
		color.Green("------------------------------------------------------")
		color.Green("  All issues seem to still have pending linked issues. ")
		color.Green("------------------------------------------------------")
	}

}
