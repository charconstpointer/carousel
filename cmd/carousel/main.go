package main

import (
	"fmt"

	"github.com/slack-go/slack"
)

func main() {
	api := slack.New("xoxb-3713913975109-3702272114647-dG31LGXHf7yr2rnsb2tnbX7h")
	users, err := api.GetUsers()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	for _, user := range users {
		fmt.Printf("%s\n", user.Name)
	}
}
