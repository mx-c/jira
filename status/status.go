package status

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	jira_api "github.com/possum3d/jira/api"
	clients_api "github.com/possum3d/jira/clients/api"
)

type TransitionsMessage struct {
	Transitions []*Transition `json:"transitions"`
}

// Recommanded: fmt.Sprintf("%s (%s)", transition.Name, transition.StatusCategory.Name)
type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   *struct {
		StatusCategory *struct {
			Name string `json:"name"`
		} `json:"statusCategory"`
	} `json:"to"`
}

type Status struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	StatusCategory *struct {
		Name string `json:"name"`
	} `json:"statusCategory"`
}

type IssueMessage struct {
	Transitions []*Transition
	Fields      *struct {
		Summary string `json:"summary"`
		Status  *Status
	} `json:"fields"`
	Key string `json:"key"`
}

func formatCurrent(issue *IssueMessage) string {
	url := fmt.Sprintf("https://mention-team.atlassian.net/browse/%s", issue.Key)
	summary := issue.Fields.Summary
	status := fmt.Sprintf(
		"%s (%s)",
		issue.Fields.Status.Name,
		issue.Fields.Status.StatusCategory.Name,
	)
	log := fmt.Sprintf("------\n%s\n%s\n%s", summary, url, status)
	return log
}

func Get(client clients_api.Client, ticket string) {
	issue := currentStatus(client, ticket)
	log := formatCurrent(issue)

	if len(issue.Transitions) > 0 {
		log = fmt.Sprintf("%s\navailable transitions:", log)
	}
	for _, trans := range issue.Transitions {

		log = fmt.Sprintf(
			"%s\n-%s (%s)",
			log,
			trans.Name,
			trans.To.StatusCategory.Name,
		)
	}

	fmt.Printf("%s\n\n", log)
	//availableTransition(client, ticket)
}

func Update(client clients_api.Client, ticket, to string) {
	issue := currentStatus(client, ticket)

	var id string

	for _, t := range issue.Transitions {
		if t.Name == to {
			id = t.ID
			break
		}
	}

	if id == "" {
		panic(fmt.Errorf("%s is not in available transitions", to))
	}

	success := updateStatus(client, ticket, id)

	if success {
		fmt.Printf("%s\nswitched to %s", formatCurrent(issue), to)
	}
}

func currentStatus(client clients_api.Client, ticket string) *IssueMessage {
	usableReq := &jira_api.SimpleRequest{
		Endpoint: fmt.Sprintf(
			"https://mention-team.atlassian.net/rest/api/2/issue/%s",
			ticket,
		),
		Querystring: url.Values{
			"expand": []string{"transitions"},
		},
		Data:   nil,
		Method: "GET",
		Header: make(http.Header),
	}
	reply, err := client.Request(context.Background(), usableReq)
	if err != nil {
		panic(err)
	}
	if reply.StatusCode != 200 {
		panic(fmt.Errorf("%s", reply.Body))
	}

	issueMsg := &IssueMessage{}
	err = json.Unmarshal(reply.Body, issueMsg)

	if err != nil {
		panic(err)
	}

	return issueMsg
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/#api-api-3-issue-issueIdOrKey-transitions-post
func updateStatus(client clients_api.Client, ticket, statusId string) bool {
	usableReq := &jira_api.SimpleRequest{
		Endpoint: fmt.Sprintf(
			"https://mention-team.atlassian.net/rest/api/3/issue/%s/transitions",
			ticket,
		),
		Querystring: url.Values{},
		Data: []byte(fmt.Sprintf(`
		{	
			"transition": {
				"id": %s
			}
		}
		`, statusId)),
		Method: "POST",
		Header: make(http.Header),
	}

	usableReq.Header.Set("Content-Type", "application/json")

	reply, err := client.Request(context.Background(), usableReq)
	if err != nil {
		panic(err)
	}

	if reply.StatusCode != 204 {
		panic(fmt.Errorf("%v %s", reply.StatusCode, reply.Body))
	}

	return true
}
