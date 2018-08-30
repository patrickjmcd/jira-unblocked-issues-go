package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/fatih/color"
)

func getUnspecifiedKey(key string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", key)
	readVal, err := reader.ReadString('\n')

	if err != nil {
		fmt.Printf("You need to specify a %s\n", key)
		os.Exit(1)
	}
	trimmedVal := strings.TrimSuffix(readVal, "\n")
	fmt.Println(trimmedVal)
	return trimmedVal
}

func getEnvVariablesOrAsk() (string, string, string) {
	jiraURL, urlExists := os.LookupEnv("JIRA_URL")
	if !urlExists {
		jiraURL = getUnspecifiedKey("Jira URL")
	}

	jiraUsername, usernameExists := os.LookupEnv("JIRA_USERNAME")
	if !usernameExists {
		jiraUsername = getUnspecifiedKey("Jira Username")
	}
	jiraPassword, passwordExists := os.LookupEnv("JIRA_PASSWORD")
	if !passwordExists {
		jiraPassword = getUnspecifiedKey("Jira Password")
	}

	return jiraURL, jiraUsername, jiraPassword
}

func main() {
	verbose := false
	if len(os.Args) < 2 {
		color.Red("Need to provide a project name as the first argument")
		os.Exit(1)
	}
	projectName := os.Args[1]

	if len(os.Args) > 2 {
		if os.Args[2] == "v" {
			verbose = true
		}
	}

	url, username, password := getEnvVariablesOrAsk()
	transport := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}
	jiraClient, err := jira.NewClient(transport.Client(), url)
	if err != nil {
		fmt.Println("Couldn't log on to the Jira server.")
		os.Exit(1)
	}

	// checkResolvedLinkedIssuesForProject(jiraClient, projectName, verbose)
	issuesWithResolved := getResolvedLinkedIssuesForProject(jiraClient, projectName, verbose)
	if len(issuesWithResolved) > 0 {
		color.Red("------------------------------------------------------")
		color.Red("   The following %d issues have completed linked issues  ", len(issuesWithResolved))
		color.Red("------------------------------------------------------")
		for _, issue := range issuesWithResolved {
			color.Red("[%s] %s", issue.Key, issue.Fields.Summary)
		}
		color.Red("------------------------------------------------------")
	} else {
		color.Green("------------------------------------------------------")
		color.Green("  All issues seem to still have pending linked issues. ")
		color.Green("------------------------------------------------------")
	}

}
