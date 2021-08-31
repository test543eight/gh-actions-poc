package cron

import (
	"context"

	"github.com/google/go-github/v37/github"
	ci "github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci"
	"github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci/pkg/bot"
	"github.com/gravitational/trace"
)

func DimissStaleWorkflowRunsForExternalContributors(token, repoName, repoOwner string, clt *github.Client) error {
	pulls, _, err := clt.PullRequests.List(context.TODO(), repoOwner, repoName, &github.PullRequestListOptions{State: ci.OPEN})
	if err != nil {
		return err
	}
	for _, pull := range pulls {
		err := bot.DismissStaleWorkflowRuns(token, *pull.Base.User.Login, *pull.Base.Repo.Name, *pull.Head.Ref, clt)
		if err != nil {
			return trace.Wrap(err)
		}

	}
	return nil
}
