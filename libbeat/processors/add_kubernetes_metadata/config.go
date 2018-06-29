// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package add_kubernetes_metadata

import (
	"time"

	"github.com/elastic/beats/libbeat/common"
)

type kubeAnnotatorConfig struct {
	InCluster  bool          `config:"in_cluster"`
	KubeConfig string        `config:"kube_config"`
	Host       string        `config:"host"`
	Namespace  string        `config:"namespace"`
	SyncPeriod time.Duration `config:"sync_period"`
	// Annotations are kept after pod is removed, until they haven't been accessed
	// for a full `cleanup_timeout`:
	CleanupTimeout time.Duration `config:"cleanup_timeout"`
	Indexers       PluginConfig  `config:"indexers"`
	Matchers       PluginConfig  `config:"matchers"`
}

type Enabled struct {
	Enabled bool `config:"enabled"`
}

type PluginConfig []map[string]common.Config

func defaultKubernetesAnnotatorConfig() kubeAnnotatorConfig {
	return kubeAnnotatorConfig{
		InCluster:      true,
		SyncPeriod:     1 * time.Second,
		CleanupTimeout: 60 * time.Second,
	}
}

func defaultKubernetesAnnotatorConfigMetricbeat() kubeAnnotatorConfig {
	config := defaultKubernetesAnnotatorConfig()
	empty := common.NewConfig()
	config.Indexers = PluginConfig{
		{
			IPPortIndexerName: *empty,
		},
	}

	//Add field matcher with field to lookup as metricset.host
	fieldCfg, err := common.NewConfigFrom(map[string]interface{}{
		"lookup_fields": []string{"metricset.host"},
	})
	if err == nil {
		config.Matchers = PluginConfig{
			{
				FieldMatcherName: *fieldCfg,
			},
		}
	}
}

func defaultKubernetesAnnotatorConfigFilebeat() kubeAnnotatorConfig {
	config := defaultKubernetesAnnotatorConfig()
	empty := common.NewConfig()
	config.Indexers = PluginConfig{
		{
			ContainerIndexerName: *empty,
		},
	}

	//Add a log path matcher which can extract container ID from the "source" field.
	fieldCfg, err := common.NewConfigFrom(config)
	if err == nil {
		config.Matchers = PluginConfig{
			{
				// XXX this symbol is only available in filebeat, this won't compile:
				LogPathMatcherName: *empty,
			},
		}
	}
}

func defaultKubernetesAnnotatorConfigPacketbeat() kubeAnnotatorConfig {
	config := defaultKubernetesAnnotatorConfig()
	empty := common.NewConfig()
	config.Indexers = PluginConfig{
		{
			IPPortIndexerName: *empty,
		},
	}

	//Add field matcher with field to lookup ip:port pairs
	formatCfg, err := common.NewConfigFrom(map[string]interface{}{
		"format": "%{[ip]}:%{[port]}",
	})
	if err == nil {
		config.Matchers = PluginConfig{
			{
				FieldFormatMatcherName: *formatCfg,
			},
		}
	}
}
