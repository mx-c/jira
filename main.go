package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	"github.com/c-bata/go-prompt"
	flags "github.com/jessevdk/go-flags"
	jira_api "github.com/possum3d/jira/api"
	"github.com/possum3d/jira/clients"
	"github.com/possum3d/jira/notifications"
	"github.com/possum3d/jira/status"
)

const (
	notifs        string = "notifs"
	ticket_status string = "status"
	status_get    string = "get"
	status_update string = "update"
	help          string = "help"
)

var opts struct {
	Interactive bool `short:"i" long:"interactive" description:"Use ./jira interactively."`
}

var ticketRegexp = regexp.MustCompile(`([A-Z]+-\d+)`)

func main() {

	conf, err := loadConfig()

	if err != nil {
		os.Exit(1)
	}
	_, err = flags.Parse(&opts)

	if err != nil {
		panic(err)
	}

	if opts.Interactive {
		interactive(conf)
	} else {
		panic("only interactive supported for now")
	}
}

func interactive(conf *jira_api.Config) {
	scrapClient := clients.MustNewScrapClient(conf)
	restClient := clients.MustNewRESTClient(conf)

	fmt.Println("Please select command:")
	t := prompt.Input("> ", commandCompleter)
	switch t {
	case notifs:
		notifications.Get(scrapClient)
		os.Exit(1)
	case ticket_status:
		fmt.Println("Ticket:")
		ticket := prompt.Input(
			"> ",
			func(d prompt.Document) []prompt.Suggest {
				return prompt.FilterHasPrefix(
					[]prompt.Suggest{
						{Text: "RRS-", Description: "rrs ticket prefix"},
						{Text: "BUGS-", Description: "bug ticket prefix"},
					},
					d.GetWordBeforeCursor(),
					true,
				)
			},
		)
		fmt.Println("Action:")
		cmd := prompt.Input("> ", StatusCommandCompleter)
		switch cmd {
		case status_get:
			status.Get(restClient, ticket)
		case status_update:
			fmt.Println("Transition to:")
			newStatus := prompt.Input("> ", UpdateStatusCompleter)
			status.Update(restClient, ticket, newStatus)
		}
		os.Exit(1)
	case help:
		helpCommand()
		os.Exit(1)
	default:
		panic(fmt.Errorf("Unknown command"))
	}
	fmt.Println("aborted command " + t)

}

func UpdateStatusCompleter(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "Start", Description: "Start"},
		{Text: "Abandoned", Description: "Abandoned"},
		{Text: "Stop", Description: "Stop"},
		{Text: "Ready for review", Description: "Ready for review"},
		{Text: "In Review", Description: "In Review"},
		{Text: "Need changes", Description: "Need changes"},
		{Text: "Ready", Description: "Ready"},
		{Text: "Reopen", Description: "Reopen"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)

}

func StatusCommandCompleter(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "get", Description: "retrieve status for reference"},
		{Text: "update", Description: "modifiy status"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func commandCompleter(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "notifs", Description: "display latest notifs"},
		{Text: "status", Description: "get and update ticket status"},
		// help about above commands
		{Text: "help", Description: "help on supported commands"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func loadConfig() (*jira_api.Config, error) {
	path := os.Getenv("HOME") + "/.jira/config.json"

	data, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		conf := jira_api.Config{
			Email:      "email",
			BasicToken: "basic token",
		}
		confBuf, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			panic(err)
		}
		return nil, fmt.Errorf(
			"Config file %v does not exist. Create it with this content:\n %s",
			path,
			string(confBuf),
		)
	}

	if err != nil {
		return nil, err
	}

	conf := &jira_api.Config{}
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("Can not load config %v: %v", err)
	}

	return conf, nil
}

// It retrieves the ticket from open branch where the jira soft is called.
func smartGetTicket() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		panic(err)
	}

	match := ticketRegexp.FindStringSubmatch(string(out))
	if len(match) < 1 {
		panic(fmt.Errorf("Failed to retrieve a ticket match"))
	}

	return match[0]
}
