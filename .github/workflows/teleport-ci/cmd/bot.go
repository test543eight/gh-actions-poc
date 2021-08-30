package main

import (
	"context"
	"flag"
	"log"
	"os"

	ci "github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci"
	"github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci/pkg/bot"
	"github.com/gravitational/gh-actions-poc/.github/workflows/teleport-ci/pkg/environment"

	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

func main() {
	var token = flag.String("token", "", "token is the Github authentication token.")
	var workflowCredentials = flag.String("workflow-creds", "", "workflow-creds is an access token with write permissions to actions.")
	var reviewers = flag.String("reviewers", "", "reviewers is a string representing a json object that maps authors to required reviewers for that author.")
	var defaultReviewers = flag.String("default-reviewers", "", "default-reviewers represents reviewers for external contributors or any author that does not have a key-value pair in '--reviewers'.")
	flag.Parse()

	if *token == "" {
		log.Fatal("missing authentication token.")
	}
	if *reviewers == "" {
		log.Fatal("missing assignments flag.")
	}
	if *defaultReviewers == "" {
		log.Fatal("missing default-reviewers flag.")

	}
	subcommand := os.Args[len(os.Args)-1]

	// Creating and authenticating the Github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Getting event object path
	path := os.Getenv(ci.GITHUBEVENTPATH)

	env, err := environment.New(environment.Config{Client: client,
		Reviewers:        *reviewers,
		DefaultReviewers: *defaultReviewers,
		EventPath:        path,
		WorkflowCreds:    *workflowCredentials,
	})
	if err != nil {
		log.Fatal(err)
	}

	bot, err := bot.New(bot.Config{Environment: env})
	if err != nil {
		log.Fatal(err)
	}
	switch subcommand {
	case ci.ASSIGN:
		log.Println("Assigning reviewers.")
		err = bot.Assign()
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Assign completed.")

	case ci.CHECK:
		log.Println("Checking reviewers.")
		err = bot.Check()
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Check completed.")
	default:
		log.Fatalf("Unknown subcommand: %v", subcommand)
	}

}
