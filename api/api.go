package api

import (
	"net/http"
	"net/url"
)

type Config struct {
	Email             string `json:"email"`
	BasicToken        string `json:"basic_token"`
	CloudId           string `json:"cloud_id"`
	CloudSessionToken string `json:"cloud_session_token"`
}

type SimpleRequest struct {
	Endpoint    string
	Querystring url.Values
	Data        []byte
	Method      string
	Header      http.Header
}

type SimpleReply struct {
	StatusCode int
	Status     string
	Body       []byte
	Header     http.Header
}
