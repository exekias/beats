package main

import (
	"github.com/elastic/beats/libbeat/plugin"
	"github.com/elastic/beats/x-pack/docker-logging-plugin/input"
)

// Bundle autodiscover provider and input for docker driver plugin
var Bundle = plugin.Bundle(map[string][]interface{}{
	"input": {input.Plugin},
})
