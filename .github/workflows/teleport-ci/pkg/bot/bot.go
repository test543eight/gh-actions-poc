package bot

import (
	"context"

	"github.com/google/go-github/v37/github"
	"github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci/pkg/environment"
	"github.com/gravitational/trace"
)

// Config is used to configure Bot
type Config struct {
	Environment *environment.Environment
}

// Bot assigns reviewers and checks assigned reviewers for a pull request
type Bot struct {
	Environment *environment.Environment
	invalidate  invalidate
	verify      verify
}

type invalidate func(string, string, string, int, []review, *github.Client) error
type verify func(string, string, string, string) error

// New returns a new instance of  Bot
func New(c Config) (*Bot, error) {
	var ch Bot
	err := c.CheckAndSetDefaults()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	ch.Environment = c.Environment
	ch.invalidate = invalidateApprovals
	ch.verify = verifyCommit
	return &ch, nil
}

// CheckAndSetDefaults verifies configuration and sets defaults
func (c *Config) CheckAndSetDefaults() error {
	if c.Environment == nil {
		return trace.BadParameter("missing parameter Environment.")
	}
	return nil
}

// invalidateApprovals dismisses all reviews on a pull request
func invalidateApprovals(repoOwner, repoName, msg string, number int, reviews []review, clt *github.Client) error {
	for _, v := range reviews {
		_, _, err := clt.PullRequests.DismissReview(context.TODO(), repoOwner, repoName, number, v.id, &github.PullRequestReviewDismissalRequest{Message: &msg})
		if err != nil {
			return trace.Wrap(err)
		}
	}
	return nil
}
