package review

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)



// PullRequestMetadata contains information about the pull request
// this is used to make api requests
type PullRequestMetadata struct {
	PullRequest struct {
		Number int `json:"number"`
		User   struct {
			Name string `json:"login"`
		}
		Head struct {
			Repo struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			} `json:"repo"`
		} `json:"head"`
	} `json:"pull_request"`
}


type review struct {
	reviewer string 
	status string 

}

// CheckReviewers checks if all the reviewers have approved a pull request
// returns nil if all required reviewers have approved, returns error if not
func CheckReviewers(path string) error {
	data, err := unmarshalPullRequestMetadata(path)
	if err != nil {
		return err
	}

	client := MakeGHClient()
	listOpts := github.ListOptions{}

	reviews, res, err := client.PullRequests.ListReviews(context.TODO(), data.PullRequest.Head.Repo.Owner.Login, data.PullRequest.Head.Repo.Name, data.PullRequest.Number, &listOpts)
	if err != nil {
		return err
	}
	log.Printf("status: %s", res.Status)



	var currentReviews []review
	for _, rev := range reviews {
		currentReviews = append(currentReviews, review{reviewer: *rev.User.Name, status: *rev.State})
	}
	// TODO: Get required reviewers
	return checkReviewers([]string{}, currentReviews)
	
}

// checkReviewers checks to see if all the required reviwers are in the current 
// reviewer slice 
func checkReviewers(required []string, current []review) error {
	// TODO: check if all required reviewers are in current 
	return nil 
}


// unmarshalPullRequestData unmarshals pull request metadata from json file given the path
func unmarshalPullRequestMetadata(path string) (PullRequestMetadata, error) {
	var metadata PullRequestMetadata

	jsonFile, err := os.Open(path)
	if err != nil {
		return PullRequestMetadata{}, err
	}
	body, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return PullRequestMetadata{}, err
	}
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return PullRequestMetadata{}, err
	}
	return metadata, nil

}

// MakeGHClient authenticates and makes the Github client
func MakeGHClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv(TOKEN)},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}


