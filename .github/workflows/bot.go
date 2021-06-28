package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (

	// ASSIGN is the argument to assign reviewers
	ASSIGN = "assign-reviewers"
	// CHECK is the argument to check reviewers
	CHECK = "check-reviewers"
)

func main() {

	args := os.Args[1:]
	if len(args) != 1 {
		panic("one argument needed \nassign-reviewers or check-reviewers")
	}

	path := os.Getenv("GITHUB_EVENT_PATH")

	switch args[0] {
	case ASSIGN:
		log.Println("Assigning...")
		reviewers := []string{}
		err := AssignReviewers(path, reviewers)
		if err != nil {
			panic(err)
		}

	case CHECK:
		log.Println("Checking...")

		reviewers := []string{}
		_, err := CheckReviewers(path, reviewers)
		if err != nil {
			panic(err)
		}

	}

}

const (
	approvedState = "APPROVED"
)

// PullRequestEventData ...
type PullRequestEventData struct {
	Number     int `json:"number"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

// PullRequestReviewEventData ...
type PullRequestReviewEventData struct {
	PullRequest struct {
		Number int `json:"number"`
		Head   struct {
			Repo struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			} `json:"repo"`
		} `json:"head"`
	} `json:"pull_request"`
}

// AssignReviewers ...
func AssignReviewers(path string, reviewerSlice []string) error {
	data, err := GetPullRequestDataFromPath(path)
	if err != nil {
		return err
	}
	client := MakeGHClient()

	reviewers := github.ReviewersRequest{Reviewers: []string{"russjones"}}

	_, res, err := client.PullRequests.RequestReviewers(context.TODO(), data.Repository.Owner.Login, data.Repository.Name, data.Number, reviewers)
	if err != nil {
		return err
	}
	log.Printf("status: %v", res.Status)
	return nil
}

// CheckReviewers ...
func CheckReviewers(path string, reviewers []string) (bool, error) {
	data, err := GetPullRequestReviewDataFromPath(path)
	if err != nil {
		return false, err
	}

	client := MakeGHClient()
	listOpts := github.ListOptions{Page: 10, PerPage: 10}

	reviews, res, err := client.PullRequests.ListReviews(context.TODO(), data.PullRequest.Head.Repo.Owner.Login, data.PullRequest.Head.Repo.Name, data.PullRequest.Number, &listOpts)
	if err != nil {
		return false, err
	}
	fmt.Println(res.Status)
	for _, rev := range reviews {
		if rev.State != nil && *rev.State != approvedState {
			return false, nil
		}
	}
	log.Printf("%+v", reviews)
	return true, nil
}

// GetPullRequestDataFromPath ...
func GetPullRequestDataFromPath(path string) (PullRequestEventData, error) {
	var pullRequestData PullRequestEventData

	jsonFile, err := os.Open(path)
	if err != nil {
		return PullRequestEventData{}, err
	}
	body, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return PullRequestEventData{}, err
	}

	err = json.Unmarshal(body, &pullRequestData)
	if err != nil {
		return PullRequestEventData{}, err
	}

	return pullRequestData, nil
}

// GetPullRequestReviewDataFromPath ...
func GetPullRequestReviewDataFromPath(path string) (PullRequestReviewEventData, error) {
	var pullRequestReviewData PullRequestReviewEventData

	jsonFile, err := os.Open(path)
	if err != nil {
		return PullRequestReviewEventData{}, err
	}
	body, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return PullRequestReviewEventData{}, err
	}
	err = json.Unmarshal(body, &pullRequestReviewData)
	if err != nil {
		return PullRequestReviewEventData{}, err
	}
	return pullRequestReviewData, nil

}

// MakeGHClient ...
func MakeGHClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
