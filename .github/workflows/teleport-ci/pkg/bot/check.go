package bot

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	ci "github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci"
	teleportci "github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci"
	"github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci/pkg/environment"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v37/github"
	"github.com/gravitational/trace"
)

// Check checks if all the reviewers have approved the pull request in the current context.
func (c *Bot) Check() error {
	env := c.Environment
	pr := c.Environment.PullRequest
	err := dismissStaleWorkflows(env.WorkflowCreds, pr.RepoOwner, pr.RepoName, pr.BranchName, env.Client)
	if err != nil {
		return trace.Wrap(err)
	}
	log.Println("Dismissed stale runs.")
	listOpts := github.ListOptions{}
	reviews, _, err := env.Client.PullRequests.ListReviews(context.TODO(), pr.RepoOwner,
		pr.RepoName,
		pr.Number,
		&listOpts)
	if err != nil {
		return trace.Wrap(err)
	}
	currentReviewsSlice := []review{}
	for _, rev := range reviews {
		currReview := review{name: *rev.User.Login, status: *rev.State, commitID: *rev.CommitID, id: *rev.ID, submittedAt: rev.SubmittedAt}
		currentReviewsSlice = append(currentReviewsSlice, currReview)
	}
	return c.check(c.Environment.IsInternal(pr.Author), pr, c.Environment.GetReviewersForAuthor(pr.Author), mostRecent(currentReviewsSlice))
}

// check checks to see if all the required reviewers have approved and invalidates
// approvals for external contributors if a new commit is pushed
func (c *Bot) check(isInternal bool, pr *environment.PullRequestMetadata, required []string, currentReviews []review) error {
	if len(currentReviews) == 0 {
		return trace.BadParameter("pull request has no reviews.")
	}
	log.Printf("checking if %v has approvals from the required reviewers %+v", pr.Author, required)
	for _, requiredReviewer := range required {
		if !containsApprovalReview(requiredReviewer, currentReviews) {
			return trace.BadParameter("all required reviewers have not yet approved.")
		}
	}

	if hasNewCommit(pr.HeadSHA, currentReviews) && !isInternal {
		// Check file changes/commit verification
		err := c.verify(pr.RepoOwner, pr.RepoName, pr.BaseSHA, pr.HeadSHA)
		if err != nil {
			if validationErr := c.invalidate(pr.RepoOwner, pr.RepoName, dismissMessage(pr, required), pr.Number, currentReviews, c.Environment.Client); validationErr != nil {
				return trace.Wrap(validationErr)
			}
			return trace.Wrap(err)
		}
	}
	return nil
}

// mostRecent returns a list of the most recent review from each required reviewer
func mostRecent(currentReviews []review) []review {
	mostRecentReviews := make(map[string]review)
	for _, rev := range currentReviews {
		val, ok := mostRecentReviews[rev.name]
		if !ok {
			mostRecentReviews[rev.name] = rev
		} else {
			setTime := val.submittedAt
			currTime := rev.submittedAt
			if currTime.After(*setTime) {
				mostRecentReviews[rev.name] = rev
			}
		}
	}
	reviews := []review{}
	for _, v := range mostRecentReviews {
		reviews = append(reviews, v)
	}
	return reviews
}

// review is a pull request review
type review struct {
	name        string
	status      string
	commitID    string
	id          int64
	submittedAt *time.Time
}

func containsApprovalReview(reviewer string, reviews []review) bool {
	for _, rev := range reviews {
		if rev.name == reviewer && rev.status == ci.APPROVED {
			return true
		}
	}
	return false
}

// dimissMessage returns the dimiss message when a review is dismissed
func dismissMessage(pr *environment.PullRequestMetadata, required []string) string {
	var buffer bytes.Buffer
	buffer.WriteString("New commit pushed, please rereview ")
	for _, reviewer := range required {
		buffer.WriteString(fmt.Sprintf("@%v ", reviewer))
	}
	return buffer.String()
}

// hasNewCommit sees if the pull request has a new commit
// by comparing commits after the push event
func hasNewCommit(headSHA string, revs []review) bool {
	for _, v := range revs {
		if v.commitID != headSHA {
			return true
		}
	}
	return false
}

// verifyCommit verfies GitHub is the commit author and that the commit is empty
func verifyCommit(repoOwner, repoName, baseSHA, headSHA string) error {
	client := github.NewClient(nil)
	comparison, _, err := client.Repositories.CompareCommits(context.TODO(), repoOwner, repoName, baseSHA, headSHA)
	if err != nil {
		return trace.Wrap(err)
	}
	if len(comparison.Files) != 0 {
		return trace.BadParameter("detected file change.")
	}
	commit, _, err := client.Repositories.GetCommit(context.TODO(), repoOwner, repoName, headSHA)
	if err != nil {
		return trace.Wrap(err)
	}
	verification := commit.Commit.Verification
	// Get commit object
	payload := *verification.Payload
	if strings.Contains(payload, teleportci.GITHUBCOMMIT) && *verification.Verified {
		return nil
	}
	return trace.BadParameter("commit is not verified and/or is not signed by GitHub.")
}

func dismissStaleWorkflows(token, owner, repoName, branch string, cl *github.Client) error {
	var targetWorkflow *github.Workflow
	workflows, _, err := cl.Actions.ListWorkflows(context.TODO(), owner, repoName, &github.ListOptions{})
	if err != nil {
		return err
	}
	for _, w := range workflows.Workflows {
		if *w.Name == ci.CHECKWORKFLOW {
			targetWorkflow = w
			break
		}
	}
	list, _, err := cl.Actions.ListWorkflowRunsByID(context.TODO(), owner, repoName, *targetWorkflow.ID, &github.ListWorkflowRunsOptions{Branch: branch})
	if err != nil {
		return err
	}
	sort.Sort(ByTime(list.WorkflowRuns))
	for i, run := range list.WorkflowRuns {
		if i == len(list.WorkflowRuns)-1 {
			break
		}
		err := deleteRun(token, owner, repoName, *run.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// deleteRun deletes a workflow run.
// Note: the go-github client library does not support this endpoint.
func deleteRun(token, owner, repo string, runID int64) error {
	// Creating and authenticating the client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := oauth2.NewClient(context.Background(), ts)
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%v", owner, repo, runID)
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return trace.Wrap(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return trace.Wrap(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	return nil
}

type ByTime []*github.WorkflowRun

func (s ByTime) Len() int {
	return len(s)
}

func (s ByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByTime) Less(i, j int) bool {
	time1 := s[i].CreatedAt
	time2 := s[j].CreatedAt
	return time1.Time.Before(time2.Time)
}
