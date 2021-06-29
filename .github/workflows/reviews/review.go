package reviews

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

const (
	approvedState = "APPROVED"
	// ASSIGNMENTS is the environment variable name that stores
	// which reviewers should be assigned to which authors
	ASSIGNMENTS = "ASSIGNMENTS"
	// TOKEN is the env variable name that stores the Github authentication token
	TOKEN = "GITHUB_TOKEN"
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

// AssignReviewers assigns reviewers to the pull request
func AssignReviewers(path string) error {
	data, err := unmarshalPullRequestMetadata(path)
	if err != nil {
		return err
	}
	// getting reviewers by user name
	requiredReviewers, err := getReviewers(data.PullRequest.User.Name)
	if err != nil {
		return err
	}
	client := MakeGHClient()

	reviewers := github.ReviewersRequest{Reviewers: requiredReviewers}

	_, res, err := client.PullRequests.RequestReviewers(context.TODO(),
		data.PullRequest.Head.Repo.Owner.Login,
		data.PullRequest.Head.Repo.Name, data.PullRequest.Number,
		reviewers)

	if err != nil {
		return err
	}
	log.Printf("status: %v", res.Status)
	return nil
}

// getReviewers gets the reviewers for the current user event
func getReviewers(user string) ([]string, error) {
	obj := os.Getenv(ASSIGNMENTS)

	if obj == "" {
		return nil, errors.New("reviewers not found")
	}
	m := make(map[string][]string)

	err := json.Unmarshal([]byte(obj), &m)
	if err != nil {
		return nil, err
	}

	value, ok := m[user]
	if !ok {
		return nil, errors.New("author not found")
	}
	return value, nil
}

// CheckReviewers checks if all the reviewers have approved a pull request
// returns nil if all required reviewers have approved, returns error if not
func CheckReviewers(path string) error {
	data, err := unmarshalPullRequestMetadata(path)
	if err != nil {
		return err
	}
	requiredReviewers, err := getReviewers(data.PullRequest.User.Name)
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

	currentReviewers := make(map[string]*github.PullRequestReview)
	for _, rev := range reviews {
		currentReviewers[*rev.User.Name] = rev
	}

	for _, reviewer := range requiredReviewers {
		prReview, ok := currentReviewers[reviewer]
		if !ok {
			return errors.New("all required reviewers have not yet reviewed")
		}
		if prReview.State != nil && *prReview.State != approvedState {
			return errors.New("all required reviewers have not yet approved")
		}
	}
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
