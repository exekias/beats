// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package api

import (
	"net/http"

	"github.com/elastic/beats/libbeat/common/reload"

	uuid "github.com/satori/go.uuid"

	"github.com/elastic/beats/libbeat/common"
)

// ConfigBlock stores a piece of config from central management
type ConfigBlock struct {
	Raw string `json:"block_yml"`
}

type ConfigList struct {
	Type   string
	Blocks []*ConfigBlock
}

// ConfigBlocks holds a list of (type, list of configs) pairs
type ConfigBlocks []ConfigList

// Config returns a common.Config object holding the config from this block
func (c *ConfigBlock) Config() (*common.Config, error) {
	return common.NewConfigWithYAML([]byte(c.Raw), "")
}

// ConfigWithMeta returns a reload.ConfigWithMeta object holding the config from this block, meta will be nil
func (c *ConfigBlock) ConfigWithMeta() (*reload.ConfigWithMeta, error) {
	config, err := c.Config()
	if err != nil {
		return nil, err
	}
	return &reload.ConfigWithMeta{
		Config: config,
	}, nil
}

// Configuration retrieves the list of configuration blocks from Kibana
func (c *Client) Configuration(accessToken string, beatUUID uuid.UUID) (ConfigBlocks, error) {
	headers := http.Header{}
	headers.Set("kbn-beats-access-token", accessToken)

	resp := struct {
		ConfigBlocks []*struct {
			Type string `json:"type"`
			Raw  string `json:"block_yml"`
		} `json:"configuration_blocks"`
	}{}
	_, err := c.request("GET", "/api/beats/agent/"+beatUUID.String()+"/configuration", nil, headers, &resp)
	if err != nil {
		return nil, err
	}

	configmap := map[string][]*ConfigBlock{}
	for _, block := range resp.ConfigBlocks {
		configmap[block.Type] = append(configmap[block.Type], &ConfigBlock{Raw: block.Raw})
	}

	res := ConfigBlocks{}
	for t, blocks := range configmap {
		res = append(res, ConfigList{
			Type:   t,
			Blocks: blocks,
		})
	}

	return res, nil
}

// ConfigBlocksEqual returns true if the given config blocks are equal, false if not
func ConfigBlocksEqual(a, b ConfigBlocks) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	// TODO implement equals check

	return false
}
