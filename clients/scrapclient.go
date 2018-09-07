package clients

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	jira_api "github.com/possum3d/jira/api"
)

type ScrapClient struct {
	config    *jira_api.Config
	Client    *http.Client
	CookieJar http.CookieJar
}

func MustNewScrapClient(config *jira_api.Config) *ScrapClient {

	// Currently, session_token expected
	if config.CloudSessionToken == "" {
		panic(fmt.Errorf("Missing session_token"))
	}

	sessionCookie := &http.Cookie{
		Name:  "cloud.session.token",
		Value: config.CloudSessionToken,
	}

	URL, err := url.Parse("https://mention-team.atlassian.net/")
	if err != nil {
		panic(err)
	}

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	cookieJar.SetCookies(URL, []*http.Cookie{sessionCookie})

	client := &http.Client{
		Jar: cookieJar,
	}

	return &ScrapClient{
		config:    config,
		Client:    client,
		CookieJar: cookieJar,
	}
}

func (r *ScrapClient) Request(
	ctx context.Context,
	usableRequest *jira_api.SimpleRequest,
) (*jira_api.SimpleReply, error) {

	path := usableRequest.Endpoint + "?" + usableRequest.Querystring.Encode()
	request, err := http.NewRequest(
		usableRequest.Method,
		path,
		bytes.NewReader(usableRequest.Data),
	)

	if err != nil {
		return nil, err
	}

	request.Header = usableRequest.Header

	resp, err := r.Client.Do(request)

	// resp can be non nil even with err not nil!
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return &jira_api.SimpleReply{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       body,
		Header:     resp.Header,
	}, nil
}
