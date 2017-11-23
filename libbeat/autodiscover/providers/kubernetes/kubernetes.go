package kubernetes

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/elastic/beats/libbeat/autodiscover"
	"github.com/elastic/beats/libbeat/autodiscover/template"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/bus"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/processors/add_kubernetes_metadata"
	"github.com/ericchiang/k8s"
)

func init() {
	autodiscover.ProviderRegistry.AddProvider("kubernetes", AutodiscoverBuilder)
}

// Provider implements autodiscover provider for kubernetes
type Provider struct {
	config        *Config
	bus           bus.Bus
	watcher       *add_kubernetes_metadata.PodWatcher
	templates     *template.Mapper
	stop          chan interface{}
	startListener bus.Listener
	stopListener  bus.Listener
}

// AutodiscoverBuilder builds and returns an autodiscover provider
func AutodiscoverBuilder(bus bus.Bus, c *common.Config) (autodiscover.Provider, error) {
	config := defaultConfig()
	err := c.Unpack(&config)
	if err != nil {
		return nil, err
	}

	mapper, err := template.NewConfigMapper(config.Templates)
	if err != nil {
		return nil, err
	}

	client, err := k8s.NewInClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get in cluster configuration: %v", err)
	}

	ctx := context.Background()
	var host string
	if config.Host != "" {
		host = config.Host
	} else {
		podName := os.Getenv("HOSTNAME")
		logp.Info("Using pod name %s and namespace %s", podName, client.Namespace)
		if podName == "localhost" {
			host = "localhost"
		} else {
			pod, error := client.CoreV1().GetPod(ctx, podName, client.Namespace)
			if error != nil {
				logp.Err("Querying for pod failed with error: ", error.Error())
				logp.Info("Unable to find pod, setting host to localhost")
				host = "localhost"
			} else {
				host = pod.Spec.GetNodeName()
			}
		}
	}

	var indexersConfig []map[string]common.Config
	for key, cfg := range add_kubernetes_metadata.Indexing.GetDefaultIndexerConfigs() {
		indexersConfig = append(indexersConfig, map[string]common.Config{key: cfg})
	}
	metaGen := add_kubernetes_metadata.NewGenDefaultMeta(config.IncludeAnnotations, config.IncludeLabels, config.ExcludeLabels)
	indexers := add_kubernetes_metadata.NewIndexers(indexersConfig, metaGen)

	watcher := add_kubernetes_metadata.NewPodWatcher(client, indexers, 10*time.Second, 10*time.Second, host)

	start := watcher.ListenStart()
	stop := watcher.ListenStop()

	return &Provider{
		config:        config,
		bus:           bus,
		templates:     mapper,
		watcher:       watcher,
		stop:          make(chan interface{}),
		startListener: start,
		stopListener:  stop,
	}, nil
}

// Start the autodiscover process
func (d *Provider) Start() {
	go d.watcher.Run()

	go func() {
		for {
			select {
			case <-d.stop:
				d.startListener.Stop()
				d.stopListener.Stop()
				return

			case event := <-d.startListener.Events():
				d.emitContainer(event, "start")

			case event := <-d.stopListener.Events():
				d.emitContainer(event, "stop")
			}
		}
	}()
}

func (d *Provider) emitContainer(event bus.Event, flag string) {
	pod, ok := event["pod"].(common.MapStr)
	if !ok {
		logp.Err("Couldn't get a pod from watcher event")
		return
	}

	var host string
	if ip, err := pod.GetValue("pod.ip"); err == nil {
		host = ip.(string)
	}

	meta := pod.Clone()

	// Emit container info
	d.publish(bus.Event{
		flag:         true,
		"host":       host,
		"kubernetes": meta,
		"meta": common.MapStr{
			"kubernetes": meta,
		},
	})
}

func (d *Provider) publish(event bus.Event) {
	// Try to match a config
	if config := d.templates.GetConfig(event); config != nil {
		event["config"] = config
	}
	d.bus.Publish(event)
}

// Stop the autodiscover process
func (d *Provider) Stop() {
	close(d.stop)
}

func (d *Provider) String() string {
	return "docker"
}
