package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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

		// Getting reviews
		resp, err := http.Get(constructRequestReviewerEndpoint(os.Getenv(repoOwner), os.Getenv(repoName), os.Getenv(pullRequestNumber)))
		if err != nil {
			log.Fatalln(err)
		}
		var revs []Review
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		err = json.Unmarshal(body, &revs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v", revs)

		for _, rev := range revs {
			i := strconv.Itoa(rev.ID)
			if err != nil {
				panic(err)
			}
			url := constructDismissEndpoint(os.Getenv(repoOwner), os.Getenv(repoName), os.Getenv(pullRequestNumber), i)
			fmt.Println("URL:>", url)

			var jsonStr = []byte(`{"message":"message"}`)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			req.Header.Set("Accept", "application/vnd.github.v3+json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv(githubToken)))
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			fmt.Println("response Status:", resp.Status)
			fmt.Println("response Headers:", resp.Header)
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
		}
	default:
		panic("invalid input")
	}
}

// Review ...
type Review struct {
	ID int `json:"id"`
}

func constructRequestReviewerEndpoint(owner, repoName, pullRequestNum string) string {
	name := strings.Split(repoName, "/")[1]
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/requested_reviewers", owner, name, pullRequestNum)
}

func getReviewersEndpoint(owner, repo, pullRequestNum string) string {
	name := strings.Split(repoName, "/")[1]
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/reviews", owner, name, pullRequestNum)

}

func constructDismissEndpoint(owner, repoName, pullRequestNum, reviewID string) string {
	// /repos/{owner}/{repo}/pulls/{pull_number}/reviews/{review_id}/dismissals
	name := strings.Split(repoName, "/")[1]
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%s/reviews/%s/dismissals", owner, name, pullRequestNum, reviewID)
}
