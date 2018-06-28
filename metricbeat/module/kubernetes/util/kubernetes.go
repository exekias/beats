package util

import (
	"sync"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/kubernetes"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/metricbeat/mb"
)

var nilEnricher = &enricher{}

// Enricher takes Kubernetes events and enrich them with k8s metadata
type Enricher interface {
	// Start will start the Kubernetes watcher on the first call, does nothing on the rest
	// errors are logged as warning
	Start()

	// Enrich the given list of events
	Enrich([]common.MapStr)
}

type kubernetesConfig struct {
	InCluster  bool          `config:"in_cluster"`
	KubeConfig string        `config:"kube_config"`
	Host       string        `config:"host"`
	SyncPeriod time.Duration `config:"sync_period"`
}

type enricher struct {
	sync.RWMutex
	metadata       map[string]common.MapStr
	index          func(common.MapStr) string
	watcher        kubernetes.Watcher
	watcherStarted bool
}

// GetWatcher initializes a kubernetes watcher with the given
// scope (node or cluster), and resource type
func GetWatcher(base mb.BaseMetricSet, resource kubernetes.Resource, nodeScope bool) (kubernetes.Watcher, error) {
	config := kubernetesConfig{
		InCluster: true,
	}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	client, err := kubernetes.GetKubernetesClient(config.InCluster, config.KubeConfig)
	if err != nil {
		return nil, err
	}

	options := kubernetes.WatchOptions{
		SyncTimeout: config.SyncPeriod,
	}

	// Watch objects in the node only
	if nodeScope {
		options.Node = kubernetes.DiscoverKubernetesNode(config.Host, config.InCluster, client)
	}

	return kubernetes.NewWatcher(client, resource, options)
}

func NewResourceMetadataEnricher(
	base mb.BaseMetricSet,
	kind string,
	resource kubernetes.Resource,
	nodeScope bool) Enricher {

	watcher, err := GetWatcher(base, resource, nodeScope)
	if err != nil {
		logp.Warn("Error initializing Kubernetes metadata enricher: %s", err)
		return nilEnricher
	}

	metaConfig := kubernetes.MetaGeneratorConfig{}
	if err := base.Module().UnpackConfig(&metaConfig); err != nil {
		logp.Warn("Error initializing Kubernetes metadata enricher: %s", err)
		return nilEnricher
	}

	metaGen := kubernetes.NewMetaGeneratorFromConfig(&metaConfig)
	enricher := BuildMetadataEnricher(watcher,
		// update
		func(m map[string]common.MapStr, r kubernetes.Resource) {
			m[r.GetMetadata().GetNamespace()+r.GetMetadata().GetName()] = metaGen.ResourceMetadata(r)
		},
		// delete
		func(m map[string]common.MapStr, r kubernetes.Resource) {
			delete(m, r.GetMetadata().GetNamespace()+r.GetMetadata().GetName())
		},
		// index
		func(e common.MapStr) string {
			return getString(e, mb.ModuleDataKey+".namespace") + getString(e, "name")
		},
	)

	return enricher
}

func getString(m common.MapStr, key string) string {
	val, err := m.GetValue(key)
	if err != nil {
		return ""
	}

	str, _ := val.(string)
	return str
}

func BuildMetadataEnricher(
	watcher kubernetes.Watcher,
	update func(map[string]common.MapStr, kubernetes.Resource),
	delete func(map[string]common.MapStr, kubernetes.Resource),
	index func(e common.MapStr) string) Enricher {

	enricher := enricher{
		metadata: map[string]common.MapStr{},
		index:    index,
		watcher:  watcher,
	}

	watcher.AddEventHandler(kubernetes.ResourceEventHandlerFuncs{
		AddFunc: func(obj kubernetes.Resource) {
			enricher.Lock()
			defer enricher.Unlock()
			update(enricher.metadata, obj)
		},
		UpdateFunc: func(obj kubernetes.Resource) {
			enricher.Lock()
			defer enricher.Unlock()
			update(enricher.metadata, obj)
		},
		DeleteFunc: func(obj kubernetes.Resource) {
			enricher.Lock()
			defer enricher.Unlock()
			delete(enricher.metadata, obj)
		},
	})

	return &enricher
}

func (m *enricher) Start() {
	if m == nilEnricher {
		return
	}

	if !m.watcherStarted {
		err := m.watcher.Start()
		if err != nil {
			logp.Warn("Error starting Kubernetes watcher: %s", err)
		}
		m.watcherStarted = true
	}
}

func (m *enricher) Enrich(events []common.MapStr) {
	if m == nilEnricher {
		return
	}

	for _, event := range events {
		if meta := m.metadata[m.index(event)]; meta != nil {
			event.DeepUpdate(common.MapStr{
				mb.ModuleDataKey: meta,
			})
		}
	}
}
