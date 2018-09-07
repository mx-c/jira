package api

import (
	"context"

	jira_api "github.com/possum3d/jira/api"
)

type Client interface {
	Request(
		context.Context,
		*jira_api.SimpleRequest,
	) (*jira_api.SimpleReply, error)
}
