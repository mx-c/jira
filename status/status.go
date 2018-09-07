package status

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	jira_api "github.com/possum3d/jira/api"
	clients_api "github.com/possum3d/jira/clients/api"
)

func Get(client clients_api.Client, ticket string) {
	availableTransition(client, ticket)
}

func availableTransition(client clients_api.Client, ticket string) {
	usableReq := &jira_api.SimpleRequest{
		Endpoint: fmt.Sprintf(
			"https://mention-team.atlassian.net/rest/api/2/issue/%s/transitions",
			ticket,
		),
		Querystring: url.Values{},
		Data:        nil,
		Method:      "GET",
		Header:      make(http.Header),
	}
	reply, err := client.Request(context.Background(), usableReq)

	if err != nil {
		panic(err)
	}

	if reply.StatusCode != 200 {
		panic(fmt.Errorf("faily reply %v", reply))
	}

	fmt.Println(string(reply.Body))
}
