package main

import (
	"context"
	"log"
	"os"

	"../pkg/assign"

	ci "../teleport-ci"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		panic("one argument needed \nassign-reviewers or check-reviewers")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv(ci.TOKEN)},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	path := os.Getenv(ci.GITHUBEVENTPATH)
	token := os.Getenv(ci.TOKEN)
	reviewers := os.Getenv("ASSIGNMENTS")

	cfg := assign.Config{
		Client:    client,
		EventPath: path,
		Token:     token,
		Reviewers: reviewers,
	}

	switch args[0] {
	case ci.ASSIGN:
		log.Println("Assigning...")

		env, err := assign.New(cfg)
		if err != nil {
			log.Fatal(err)
		}
		err = env.Assign()
		if err != nil {
			log.Fatal(err)
		}

	case ci.CHECK:
		log.Println("Checking...")

		// err := reviews.CheckReviewers(path)
		// if err != nil {
		// 	panic(err)
		// }
	}

}
