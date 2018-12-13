package autodiscover

import (
	"github.com/docker/go-plugins-helpers/sdk"
	"github.com/gofrs/uuid"

	"github.com/elastic/beats/libbeat/autodiscover/builder"

	"github.com/elastic/beats/libbeat/autodiscover"
	"github.com/elastic/beats/libbeat/autodiscover/providers"
	"github.com/elastic/beats/libbeat/autodiscover/template"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/bus"
	"github.com/elastic/beats/libbeat/common/cfgwarn"
	"github.com/elastic/beats/libbeat/logp"
)

// Plugin for dockerplugin type input for Filebeat
var Plugin = providers.Plugin("dockerplugin", Builder)

// Provider implements autodiscover provider for docker containers
type Provider struct {
	config    *Config
	bus       bus.Bus
	uuid      uuid.UUID
	builders  autodiscover.Builders
	appenders autodiscover.Appenders
	templates template.Mapper
	handler   sdk.Handler
}

func Builder(bus bus.Bus, uuid uuid.UUID, c *common.Config) (autodiscover.Provider, error) {
	cfgwarn.Experimental("The docker plugin autodiscover is enabled")
	config := defaultConfig()
	err := c.Unpack(&config)
	if err != nil {
		return nil, err
	}
	mapper, err := template.NewConfigMapper(config.Templates)
	if err != nil {
		return nil, err
	}
	builders, err := autodiscover.NewBuilders(config.Builders, config.HintsEnabled)
	if err != nil {
		return nil, err
	}
	appenders, err := autodiscover.NewAppenders(config.Appenders)
	if err != nil {
		return nil, err
	}
	provider := &Provider{
		config:    config,
		bus:       bus,
		uuid:      uuid,
		builders:  builders,
		appenders: appenders,
		templates: mapper,
		handler:   sdk.NewHandler(`{"Implements": ["LoggingDriver"]}`),
	}
	initHandlers(provider)
	return provider, nil
}

// Start the autodiscover process
func (d *Provider) Start() {
	go func() {
		// Listen to unix socket, handlers will publish events
		if err := d.handler.ServeUnix("filebeat", 0); err != nil {
			logp.Error(err)
		}
	}()
}
func (d *Provider) publish(event bus.Event) {
	// try builders:
	hints := d.generateHints(event)
	if config := d.builders.GetConfig(hints); config != nil {
		event["config"] = config
	}
	// Call all appenders to append any extra configuration
	d.appenders.Append(event)

	event["provider"] = d.uuid
	d.bus.Publish(event)
}
func (d *Provider) generateHints(event bus.Event) bus.Event {
	// Try to build a config with enabled builders. Send a provider agnostic payload.
	// Builders are Beat specific.
	e := bus.Event{}
	var dockerMeta common.MapStr
	if rawDocker, err := common.MapStr(event).GetValue("docker.container"); err == nil {
		dockerMeta = rawDocker.(common.MapStr)
		e["container"] = dockerMeta
	}
	if path, ok := event["path"]; ok {
		e["path"] = path
	}
	if labels, err := dockerMeta.GetValue("labels"); err == nil {
		l := labels.(common.MapStr)
		hints := builder.GenerateHints(l, "", d.config.Prefix)
		e["hints"] = hints
	}
	return e
}

// Stop the autodiscover process
func (d *Provider) Stop() {}
func (d *Provider) String() string {
	return "dockerprovider"
}
