package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	pullRequestNumber = "PULL_REQUEST_NUMBER"
	pullRequestAuthor = "PULL_REQUEST_AUTHOR"
	repoOwner         = "REPOSITORY_OWNER"
	repoName          = "REPOSITORY_NAME"
	githubToken       = "GITHUB_TOKEN"
	// ASSIGN is the argument to assign reviewers
	ASSIGN = "assign-reviewers"
	// CHECK is the argument to check reviewers
	CHECK = "check-reviewers"
	// DISMISS is the argument to DISMISS reviewers
	DISMISS = "dismiss-reviewers"
)

func main() {
	// path := os.Getenv("GITHUB_EVENT_PATH")
	// jsonFile, err := os.Open(path)
	// if err != nil {
	// 	panic(err)
	// }
	// b, err := ioutil.ReadAll(jsonFile)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Print(string(b))
	// var url string

	// reviewers := map[string][]string{
	// 	"quinqu":    []string{"0xblush", "russjones"},
	// 	"russjones": []string{"0xblush", "quinqu"},
	// }

	args := os.Args[1:]
	if len(args) != 1 {
		panic("exactly one argument needed \nassign-reviewers or check-reviewers")
	}

	switch args[0] {
	case ASSIGN:
		fmt.Println("Assigning....")
	case CHECK:
		fmt.Println("Checking...")
	case DISMISS:
		fmt.Println("Dimissing")
	default:
		panic("invalid input")
	}
}

func constructRequestReviewerEndpoint(owner, repoName, pullRequestNum string) string {
	name := strings.Split(repoName, "/")[1]
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/requested_reviewers", owner, name, pullRequestNum)
}

func getReviewersEndpoint(owner, repo, pullRequestNum string) string {
	name := strings.Split(repoName, "/")[1]
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/reviews", owner, name, pullRequestNum)

}

func constructDismissEndpoint() string {
	return ""
}
