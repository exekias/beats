package api

import (
	"github.com/elastic/beats/libbeat/common"
	uuid "github.com/satori/go.uuid"
)

// Enroll a beat in central management, this call returns a valid access token to retrieve configurations
func (c *Client) Enroll(beatType, hostname string, beatUUID uuid.UUID, enrollmentToken string) (string, error) {
	params := common.MapStr{
		"type":             beatType,
		"host_name":        hostname,
		"enrollment_token": enrollmentToken,
	}

	resp := struct {
		AccessToken string `json:"access_token"`
	}{}

	_, err := c.request("POST", "/api/beats/agent/"+beatUUID.String(), params, &resp)
	if err != nil {
		return "", err
	}

	return resp.AccessToken, err
}
