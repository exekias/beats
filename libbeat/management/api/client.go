package api

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/setup/kibana"
	"github.com/pkg/errors"
)

const defaultTimeout = 10 * time.Second

// Client to Central Management API
type Client struct {
	client *kibana.Client
}

// ConfigFromURL generates a full kibana client config from an URL
func ConfigFromURL(kibanaURL string) (*common.Config, error) {
	kibana, err := url.Parse(kibanaURL)
	if err != nil {
		return nil, err
	}

	var username, password string
	if kibana.User != nil {
		username = kibana.User.Username()
		password, _ = kibana.User.Password()
	}
	config, err := common.NewConfigFrom(struct {
		Protocol string             `config:"protocol"`
		Host     string             `config:"host"`
		Path     string             `config:"path"`
		Username string             `config:"username"`
		Password string             `config:"password"`
		TLS      *outputs.TLSConfig `config:"ssl"`
		Timeout  time.Duration      `config:"timeout"`
	}{
		Protocol: kibana.Scheme,
		Host:     kibana.Host,
		Path:     kibana.Path,
		Username: username,
		Password: password,
		Timeout:  defaultTimeout,
	})
	if err != nil {
		return nil, err
	}

	return config, err
}

// NewClient creates and returns a kibana client
func NewClient(cfg *common.Config) (*Client, error) {
	client, err := kibana.NewKibanaClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
	}, nil
}

// do a request to the API and unmarshall the message, error if anything fails
func (c *Client) request(method, extraPath string,
	params common.MapStr, message interface{}) (int, error) {

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return 400, err
	}

	statusCode, result, err := c.client.Request(method, extraPath, nil, bytes.NewBuffer(paramsJSON))
	if err != nil {
		return statusCode, err
	}

	if statusCode >= 300 {
		err = extractError(result)
	} else {
		if err = json.Unmarshal(result, message); err != nil {
			return statusCode, errors.Wrap(err, "error unmarshaling response")
		}
	}

	return statusCode, err
}

func extractError(result []byte) error {
	var kibanaResult struct {
		Message string
	}
	if err := json.Unmarshal(result, &kibanaResult); err != nil {
		return errors.Wrap(err, "parsing kibana response")
	}
	return errors.New(kibanaResult.Message)
}
