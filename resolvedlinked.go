package main

import (
	"fmt"
	"log"

	jira "github.com/andygrunwald/go-jira"
	"github.com/fatih/color"
)

func checkLinkedIssueStatus(jiraClient *jira.Client, issue *jira.Issue, c chan string, verbose bool) {
	issueLinks := issue.Fields.IssueLinks
	var linkedIssues []*jira.Issue
	for _, linked := range issueLinks {
		if linked.OutwardIssue != nil {
			linkedIssues = append(linkedIssues, linked.OutwardIssue)
		}
		if linked.InwardIssue != nil {
			linkedIssues = append(linkedIssues, linked.InwardIssue)
		}
	}
	linkedIssuesStillPending := false
	for _, lIssue := range linkedIssues {
		if verbose {
			switch lIssue.Fields.Status.Name {
			case "In Progress":
				color.Set(color.FgBlue)
			case "To Do":
				color.Set(color.FgGreen)
			default:
				color.Set(color.FgRed)
			}

			fmt.Printf(" -- [%s] %s = %+v\n", lIssue.Key, lIssue.Fields.Summary, lIssue.Fields.Status.Name)
			color.Unset()
		}
		if lIssue.Fields.Status.Name == "In Progress" || lIssue.Fields.Status.Name == "To Do" {
			linkedIssuesStillPending = true
		}
	}
	if !linkedIssuesStillPending && len(linkedIssues) > 0 {
		c <- fmt.Sprintf("[%s] %s", issue.Key, issue.Fields.Summary)
	}
}

func getLinkedIssuesForIssue(jiraClient *jira.Client, issue *jira.Issue) []*jira.Issue {
	issueLinks := issue.Fields.IssueLinks
	var linkedIssues []*jira.Issue
	for _, linked := range issueLinks {
		if linked.OutwardIssue != nil {
			linkedIssues = append(linkedIssues, linked.OutwardIssue)
		}
		if linked.InwardIssue != nil {
			linkedIssues = append(linkedIssues, linked.InwardIssue)
		}
	}
	return linkedIssues
}

func checkResolvedLinkedIssuesForProject(jiraClient *jira.Client, projectName string, verbose bool) {
	c := make(chan string)

	searchOpts := jira.SearchOptions{
		MaxResults: 999,
	}

	projectIssues, _, pErr := jiraClient.Issue.Search("project="+projectName+" and resolved is EMPTY", &searchOpts)
	if pErr != nil {
		log.Fatal(pErr)
	}

	for _, issue := range projectIssues {
		go func(issue jira.Issue) {
			checkLinkedIssueStatus(jiraClient, &issue, c, verbose)
		}(issue)
	}

	for l := range c {
		color.Red(l)
	}

}

func getResolvedLinkedIssuesForProject(jiraClient *jira.Client, projectName string, verbose bool) []jira.Issue {
	searchOpts := jira.SearchOptions{
		MaxResults: 999,
	}

	var issuesWithResolvedLinkedIssues []jira.Issue

	projectIssues, _, pErr := jiraClient.Issue.Search("project="+projectName+" and resolved is EMPTY", &searchOpts)

	if pErr != nil {
		log.Fatal(pErr)
	}

	for _, issue := range projectIssues {
		linkedIssues := getLinkedIssuesForIssue(jiraClient, &issue)
		if verbose {
			fmt.Printf("\n[%s] %s -- %d issues\n", issue.Key, issue.Fields.Summary, len(linkedIssues))
		}
		linkedIssuesStillPending := false
		for _, lIssue := range linkedIssues {
			if verbose {
				switch lIssue.Fields.Status.Name {
				case "In Progress":
					color.Set(color.FgBlue)
				case "To Do":
					color.Set(color.FgGreen)
				default:
					color.Set(color.FgRed)
				}
				fmt.Printf(" -- [%s] %s = %+v\n", lIssue.Key, lIssue.Fields.Summary, lIssue.Fields.Status.Name)
				color.Unset()
			}
			if lIssue.Fields.Status.Name == "To Do" || lIssue.Fields.Status.Name == "In Progress" {
				linkedIssuesStillPending = true
			}
		}
		if !linkedIssuesStillPending && len(linkedIssues) > 0 {
			issuesWithResolvedLinkedIssues = append(issuesWithResolvedLinkedIssues, issue)
		}

	}
	if verbose {
		fmt.Println()
		fmt.Println()
		fmt.Println()
	}
	return issuesWithResolvedLinkedIssues
}
