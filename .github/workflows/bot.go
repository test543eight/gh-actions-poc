package main

import (
	"./reviews"
	"log"
	"os"
)

const (

	// ASSIGN is the argument to assign reviewers
	ASSIGN = "assign-reviewers"
	// CHECK is the argument to check reviewers
	CHECK = "check-reviewers"

	// GITHUBEVENTPATH is the envvariable name that
	// has the path to the event payload json file
	GITHUBEVENTPATH = "GITHUB_EVENT_PATH"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		panic("one argument needed \nassign-reviewers or check-reviewers")
	}

	path := os.Getenv(GITHUBEVENTPATH)

	switch args[0] {
	case ASSIGN:
		log.Println("Assigning...")

		err := reviews.AssignReviewers(path)
		if err != nil {
			panic(err)
		}

	case CHECK:
		log.Println("Checking...")

		err := reviews.CheckReviewers(path)
		if err != nil {
			panic(err)
		}
	}

}
