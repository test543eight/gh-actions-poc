package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/gravitational/gh-actions-poc/tool/ci"
	"github.com/gravitational/gh-actions-poc/tool/ci/pkg/bot"
	bots "github.com/gravitational/gh-actions-poc/tool/ci/pkg/bot"
	"github.com/gravitational/gh-actions-poc/tool/ci/pkg/environment"
	"github.com/gravitational/trace"

	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

func main() {
	var token = flag.String("token", "", "token is the Github authentication token.")
	var reviewers = flag.String("reviewers", "", "reviewers is a string representing a json object that maps authors to required reviewers for that author.")
	flag.Parse()

	subcommand := os.Args[len(os.Args)-1]
	ctx := context.Background()

	switch subcommand {
	case ci.Assign:
		log.Println("Assigning reviewers")
		bot, err := constructBot(ctx, *token, *reviewers)
		if err != nil {
			log.Fatal(err)
		}
		err = bot.Assign(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Assign completed")
	case "assign-reviewers-ex":
		log.Println("Assigning for external")
		err := triggerAssign(ctx, *token)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Assigning for external completed.")
	case ci.Check:
		log.Println("Checking reviewers")
		bot, err := constructBot(ctx, *token, *reviewers)
		if err != nil {
			log.Fatal(err)
		}
		err = bot.Check(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Check completed")
	case ci.Dismiss:
		log.Println("Dismissing stale runs")
		err := dismissRuns(ctx, *token)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Stale workflow run removal completed")
	default:
		log.Fatalf("Unknown subcommand: %v.\nThe following subcommands are supported:\n"+
			"\tassign-reviewers \n\t assigns reviewers to a pull request.\n"+
			"\tcheck-reviewers \n\t checks pull request for required reviewers.\n"+
			"\tdismiss-runs \n\t dismisses stale workflow runs for external contributors.\n", subcommand)
	}

}

func triggerAssign(ctx context.Context, token string) error {
	var assignTarget *github.Workflow
	clt := makeGithubClient(ctx, token)
	repository := os.Getenv(ci.GithubRepository)
	if repository == "" {
		return trace.BadParameter("environment variable GITHUB_REPOSITORY is not set")
	}
	metadata := strings.Split(repository, "/")
	if len(metadata) != 2 {
		return trace.BadParameter("environment variable GITHUB_REPOSITORY is not in the correct format,\n the valid format is '<repo owner>/<repo name>'")
	}
	workflows, _, err := clt.Actions.ListWorkflows(ctx, metadata[0], metadata[1], &github.ListOptions{})
	if err != nil {
		return trace.Wrap(err)
	}
	for _, w := range workflows.Workflows {
		log.Println(*w.Name)
		log.Println(*w.Path)
		log.Println(*w.ID)
		if *w.Name == "Assign-Target" {
			assignTarget = w
		}

	}
	pulls, _, err := clt.PullRequests.List(ctx, metadata[0], metadata[1], &github.PullRequestListOptions{State: ci.Open})
	if err != nil {
		return err
	}
	for _, pull := range pulls {
		resp, err := clt.Actions.CreateWorkflowDispatchEventByID(ctx, metadata[0], metadata[1], *assignTarget.ID, github.CreateWorkflowDispatchEventRequest{Ref: *pull.Head.SHA})
		if err != nil {
			return err
		}
		log.Printf("%+v", resp)
	}
	return nil
}

func constructBot(ctx context.Context, token, reviewers string) (*bots.Bot, error) {
	path := os.Getenv(ci.GithubEventPath)
	env, err := environment.New(environment.Config{Client: makeGithubClient(ctx, token),
		Reviewers: reviewers,
		EventPath: path,
		Token:     token,
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
	// Creating and authenticating the Github client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
