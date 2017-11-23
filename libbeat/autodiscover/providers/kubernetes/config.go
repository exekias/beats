package kubernetes

import (
	"github.com/elastic/beats/libbeat/autodiscover/template"
)

// Config for kubernetes autodiscover provider
type Config struct {
	Templates          template.MapperSettings `config:"templates"`
	IncludeLabels      []string                `config:"include_labels"`
	ExcludeLabels      []string                `config:"exclude_labels"`
	IncludeAnnotations []string                `config:"include_annotations"`
	Host               string                  `config:"host"`
}

func defaultConfig() *Config {
	return &Config{}
}
