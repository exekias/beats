package hints

import "github.com/elastic/beats/libbeat/common"

type config struct {
	Key    string         `config:"key"`
	Config *common.Config `config:"config"`
}

func defaultConfig() *config {
	return &config{
		Key: "logs",
	}
}

func (c *config) GetConfig() *common.Config {
	if c.Config != nil {
		return c.Config
	}

	// default:
	rawCfg := map[string]interface{}{
		"type": "docker",
		"containers": map[string]interface{}{
			"ids": []string{
				"${data.container.id}",
			},
		},
	}
	cfg, _ := common.NewConfigFrom(rawCfg)
	return cfg
}
