package clients

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	jira_api "github.com/possum3d/jira/api"
)

type RESTClient struct {
	config    *jira_api.Config
	basicAuth string
}

func MustNewRESTClient(config *jira_api.Config) *RESTClient {
	return &RESTClient{
		config:    config,
		basicAuth: basicAuth(config.Email, config.BasicToken),
	}
}

func (r *RESTClient) Request(
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

	request.Header.Add(
		"Authorization",
		fmt.Sprintf("Basic %s", r.basicAuth),
	)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(request)

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

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
