package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gravitational/gh-actions-poc/tool/ci"
	"github.com/gravitational/gh-actions-poc/tool/ci/pkg/bot"
	bots "github.com/gravitational/gh-actions-poc/tool/ci/pkg/bot"
	"github.com/gravitational/gh-actions-poc/tool/ci/pkg/environment"
	"github.com/gravitational/trace"

	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

const usage = "The following subcommands are supported:\n" +
	"\tassign-reviewers \n\t assigns reviewers to a pull request.\n" +
	"\tcheck-reviewers \n\t checks pull request for required reviewers.\n" +
	"\tdismiss-runs \n\t dismisses stale workflow runs for external contributors.\n"

func main() {
	var token = flag.String("token", "", "token is the Github authentication token.")
	var reviewers = flag.String("reviewers", "", "reviewers is a string representing a json object that maps authors to required reviewers for that author.")
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatalf("Subcommand required. %s\n", usage)
	}
	subcommand := os.Args[len(os.Args)-1]
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	switch subcommand {
	case ci.AssignSubcommand:
		log.Println("Assigning reviewers")
		bot, err := constructBot(ctx, *token, *reviewers)
		if err != nil {
			log.Fatal(err)
		}
		err = bot.Assign(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Assign completed.")
	case ci.CheckSubcommand:
		log.Println("Checking reviewers.")
		bot, err := constructBot(ctx, *token, *reviewers)
		if err != nil {
			log.Fatal(err)
		}
		err = bot.Check(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Check completed.")
	case ci.AssignExternalSubcommand:
		log.Println("Assigning reviewers for external contributors.")
		err := assignExternal(ctx, *token)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Assigning for external completed.")
	case ci.Dismiss:
		log.Println("Dismissing stale runs.")
		err := dismissRuns(ctx, *token)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Stale workflow run removal completed.")
	default:
		log.Fatalf("Unknown subcommand: %v.\n%s", subcommand, usage)
	}

}

func assignExternal(ctx context.Context, token string) error {
	clt := makeGithubClient(ctx, token)
	repository := os.Getenv(ci.GithubRepository)
	if repository == "" {
		return trace.BadParameter("environment variable GITHUB_REPOSITORY is not set")
	}
	metadata := strings.Split(repository, "/")
	if len(metadata) != 2 {
		return trace.BadParameter("environment variable GITHUB_REPOSITORY is not in the correct format,\n the valid format is '<repo owner>/<repo name>'")
	}

	pulls, err := getPulls(ctx, clt, metadata[0], metadata[1])
	if err != nil {
		return err 
	}
	for _, pull := range pulls {
		// Create new bot for each pull 
		env := &environment.Environment{Client: clt, PullRequest: pull, Reviewers: map[string][]string{"": {"quinqu"}}}
		bot := &bots.Bot{Environment: env}
		err := bot.Assign(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func constructBot(ctx context.Context, token, reviewers string) (*bots.Bot, error) {
	path := os.Getenv(ci.GithubEventPath)
	env, err := environment.New(environment.Config{Client: makeGithubClient(ctx, token),
		Reviewers: reviewers,
		EventPath: path,
		Token:     token,
		Context:   ctx,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	bot, err := bots.New(bots.Config{Environment: env})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return bot, nil
}

func getPulls(ctx context.Context, gc *github.Client, repoOwner, repoName string) ([]*environment.PullRequestMetadata, error) {
	var pullRequests []*environment.PullRequestMetadata
	pulls, _, err := gc.PullRequests.List(ctx, repoOwner, repoName, &github.PullRequestListOptions{State: ci.Open})
	if err != nil {
		return nil, err
	}
	for _, p := range pulls {
		prMetadata, err := toPullRequestMetadata(p)
		if err != nil {
			return nil, err
		}
		pullRequests = append(pullRequests, prMetadata)
	}
	return pullRequests, nil 
}

func toPullRequestMetadata(pull *github.PullRequest) (*environment.PullRequestMetadata, error) {
	switch {
	case pull.User.Name == nil:
		// TODO check fields 
	}

	return &environment.PullRequestMetadata{
		Author:     *pull.User.Name,
		Number:     *pull.Number,
		RepoName:   *pull.Base.Repo.Name,
		RepoOwner:  *pull.Base.User.Name,
		HeadSHA:    *pull.Head.SHA,
		BaseSHA:    *pull.Base.SHA,
		BranchName: *pull.Head.Ref,
	}, nil
}
func dismissRuns(ctx context.Context, token string) error {
	repository := os.Getenv(ci.GithubRepository)
	if repository == "" {
		return trace.BadParameter("environment variable GITHUB_REPOSITORY is not set")
	}
	metadata := strings.Split(repository, "/")
	if len(metadata) != 2 {
		return trace.BadParameter("environment variable GITHUB_REPOSITORY is not in the correct format,\n the valid format is '<repo owner>/<repo name>'")
	}
	clt := makeGithubClient(ctx, token)
	githubClient := bot.GithubClient{Client: clt}
	err := githubClient.DimissStaleWorkflowRunsForExternalContributors(ctx, token, metadata[0], metadata[1])
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func makeGithubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
