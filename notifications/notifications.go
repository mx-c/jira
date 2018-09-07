package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	jira_api "github.com/possum3d/jira/api"
	clients_api "github.com/possum3d/jira/clients/api"
)

var tplRegexp = regexp.MustCompile(`({[^}]+})`)
var tplUserRegexp = regexp.MustCompile(`({user\d+})`)
var keyUserRegexp = regexp.MustCompile(`^{user\d+}$`)

type User struct {
	Name string `json:"name"`
}

type NotificationsMessage struct {
	//Data []map[string]json.RawMessage `json:"data"`
	Data []*Notification `json:"data"`
}

type Notification struct {
	ID             string `json:"id"`
	NotificationId string `json:"notification_id"`
	EventType      string `json:"event_type"`
	Template       string `json:"template"`
	Metadata       *struct {
		Issue *struct {
			Summary string `json:"summary"`
			Url     string `json:"url"`
		} `json:"issue"`
		Content *struct {
			Url   string `json:"url"`
			Title string `json:"title"`
		}
		User *struct {
			Name string `json:"name"`
		} `json:"user"`

		User1 *struct {
			Name string `json:"name"`
		} `json:"user1"`
	} `json:"metadata"`
	ReadState string `json:"readState"`
}

type Log struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Url     string `json:"url"`
	Summary string `json:"summary"`
}

func Get(client clients_api.Client) {
	logs := retrieve(client)
	if len(logs) == 0 {
		fmt.Println("no notifications on JIRA.")
	} else {
		for _, log := range logs {
			fmt.Printf(
				"------\n%s\n%s\n%s\n\n",
				log.Message,
				log.Url,
				log.Summary,
			)
		}
		clear(client, logs[0].ID)
	}
}

func retrieve(client clients_api.Client) []*Log {
	usableReq := &jira_api.SimpleRequest{
		Endpoint: "https://mention-team.atlassian.net/gateway/api/notification-log/api/2/notifications",
		Querystring: url.Values{
			"cloudId":        []string{"b12da171-6005-4934-98f0-7d81880a70ab"},
			"direct":         []string{"true"},
			"includeContent": []string{"true"},
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
		panic(fmt.Errorf("faily reply %v", reply))
	}

	return createLogs(reply.Body)
}

func clear(client clients_api.Client, offset string) {
	usableReq := &jira_api.SimpleRequest{
		Endpoint: "https://mention-team.atlassian.net/gateway/api/notification-log/api/notifications/clearDirectUnseenNotificationCount",
		Querystring: url.Values{
			"cloudId": []string{"b12da171-6005-4934-98f0-7d81880a70ab"},
			"offset":  []string{offset},
		},
		Data:   nil,
		Method: "POST",
		Header: make(http.Header),
	}

	_, err := client.Request(context.Background(), usableReq)

	if err != nil {
		panic(err)
	}
}

func createLogs(body []byte) []*Log {
	notifMsg := &NotificationsMessage{}
	err := json.Unmarshal(body, notifMsg)
	if err != nil {
		panic(err)
	}
	var logs []*Log

	for _, notif := range notifMsg.Data {
		if notif.ReadState == "read" {
			break
		}
		msg := notif.Template
		if notif.Metadata.User != nil {
			msg = strings.Replace(msg, `{user}`, notif.Metadata.User.Name, -1)
		}
		if notif.Metadata.User1 != nil {
			msg = strings.Replace(msg, `{user1}`, notif.Metadata.User1.Name, -1)
		}

		log := &Log{
			ID:      notif.ID,
			Message: msg,
		}

		switch true {
		case notif.Metadata.Issue != nil:
			log.Url = notif.Metadata.Issue.Url
			log.Summary = notif.Metadata.Issue.Summary
		case notif.Metadata.Content != nil:
			log.Url = notif.Metadata.Content.Url
			log.Summary = notif.Metadata.Content.Title
		}

		logs = append(
			logs,
			log,
		)

	}

	return logs
}

func ExtractMessage(notifs map[string]json.RawMessage) string {

	templateMsg, ok := notifs["template"]
	if !ok {
		return ""
	}

	var template string
	err := json.Unmarshal(templateMsg, &template)

	if err != nil {
		return ""
	}

	users := make(map[string]*User)

	for key, dump := range notifs {
		log.Println(key)
		if keyUserRegexp.MatchString(key) {
			panic(fmt.Sprintf("found %s", key))
			user := &User{}
			err := json.Unmarshal(dump, user)
			if err != nil {
				continue
			}
			if user.Name == "" {
				continue
			}
			users[key] = user
		}
	}

	for key, user := range users {
		template = strings.Replace(key, user.Name, template, -1)
	}
	return template
}
