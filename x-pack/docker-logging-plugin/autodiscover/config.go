package autodiscover

import (
	"github.com/elastic/beats/libbeat/autodiscover/template"
	"github.com/elastic/beats/libbeat/common"
)

// Config for docker autodiscover provider
type Config struct {
	Prefix       string                  `config:"prefix"`
	HintsEnabled bool                    `config:"hints.enabled"`
	Builders     []*common.Config        `config:"builders"`
	Appenders    []*common.Config        `config:"appenders"`
	Templates    template.MapperSettings `config:"templates"`
}

func defaultConfig() *Config {
	return &Config{
		Prefix: "co.elastic",
	}
}
