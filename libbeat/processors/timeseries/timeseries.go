package timeseries

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/mitchellh/hashstructure"

	"github.com/elastic/beats/libbeat/asset"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/processors"
)

type timeseriesProcessor struct {
	dimensions map[string]interface{}
}

// NewTimeSeriesProcessor returns a processor to add timeseries info to events
// Events are processed to extract all their dimensions (keyword fields that
// hold a dimension of the metrics) and compute a hash of all their values into
// `ts.id` field.
func NewTimeSeriesProcessor(beatName string) (processors.Processor, error) {
	fieldsYAML, err := asset.GetFields(beatName)
	if err != nil {
		return nil, err
	}

	fields, err := common.NewFieldsFromYAML(fieldsYAML)
	if err != nil {
		return nil, err
	}

	dimensions := map[string]interface{}{}
	populateDimensions("", dimensions, fields)

	return &timeseriesProcessor{dimensions: dimensions}, nil
}

func (t *timeseriesProcessor) Run(event *beat.Event) (*beat.Event, error) {
	if event.TimeSeries && len(t.dimensions) > 0 {
		instanceFields := common.MapStr{}

		// map all dimensions & values
		for k, v := range event.Fields.Flatten() {
			if t.isDimension(k) {
				instanceFields[k] = v
			}
		}

		h, err := hashstructure.Hash(instanceFields, nil)
		if err != nil {
			// this should not happen, keep the event in any case
			return event, err
		})
		event.Fields["ts"] = common.MapStr{
			"instance": h,
		}
	}

	return event, nil
}

func (t *timeseriesProcessor) isDimension(field string) bool {
	// TODO what about objects (ie prometheus.labels)? check by prefix
	_, ok := t.dimensions[field]
	return ok
}

// put all dimension fields in the given map
func populateDimensions(prefix string, dimensions map[string]interface{}, fields common.Fields) {
	for _, f := range fields {
		name := f.Name
		if prefix != "" {
			name = prefix + "." + name
		}

		if len(f.Fields) > 0 {
			populateDimensions(name, dimensions, f.Fields)
			continue
		}

		if f.Dimension == nil {
			if f.Type == "keyword" {
				dimensions[name] = nil
			}
		} else if *f.Dimension {
			dimensions[name] = nil
		}
	}
}

func (t *timeseriesProcessor) String() string {
	return "timeseries"
}
