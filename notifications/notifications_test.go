package notifications

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractMessage(t *testing.T) {
	body, err := ioutil.ReadFile("notifications.json")
	if err != nil {
		panic(err)
	}

	logs := createLogs(body)
	assert.Equal(t, 1, len(logs))

	log := logs[0]
	assert.Equal(t, "KmeCnin mentioned you on an issue", log.Message)
}
