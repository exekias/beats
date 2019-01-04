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

// Code generated by beats/dev-tools/cmd/asset/asset.go - DO NOT EDIT.

package docker

import (
	"github.com/elastic/beats/libbeat/asset"
)

func init() {
	if err := asset.SetFields("metricbeat", "docker", Asset); err != nil {
		panic(err)
	}
}

// Asset returns asset data
func Asset() string {
	return "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"
}
