package input

import (
	"context"
	"encoding/binary"
	"io"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/plugins/logdriver"
	protoio "github.com/gogo/protobuf/io"
	"github.com/pkg/errors"
	"github.com/tonistiigi/fifo"

	"github.com/elastic/beats/filebeat/channel"
	"github.com/elastic/beats/filebeat/harvester"
	"github.com/elastic/beats/filebeat/input"
	"github.com/elastic/beats/filebeat/input/file"
	"github.com/elastic/beats/filebeat/util"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/cfgwarn"
	"github.com/elastic/beats/libbeat/logp"
)

// Plugin for dockerplugin type input for Filebeat
var Plugin = input.Plugin("dockerplugin", NewInput)

// Input is a input as docker plugin. This input is only minded to be run as a
// logging plugin for Docker. Docker handles start/stop. Events are routed through
// autodiscover like this:
//
// StartLogging -> autodiscover event -> new dockerplugin input
// Container starts
// Input reads all logs
// Container stops
// StopLogging -> autodiscover event -> stop dockerplugin input
//
type Input struct {
	started  bool
	outlet   channel.Outleter
	config   config
	cfg      *common.Config
	registry *harvester.Registry
	fifo     io.ReadWriteCloser
}

// NewInput creates a new docker input
func NewInput(
	cfg *common.Config,
	outletFactory channel.Connector,
	ctx input.Context,
) (input.Input, error) {
	cfgwarn.Experimental("Docker plugin input is enabled.")
	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, errors.Wrap(err, "reading docker input config")
	}
	outlet, err := outletFactory(cfg, ctx.DynamicFields)
	if err != nil {
		return nil, err
	}
	f, err := fifo.OpenFifo(context.Background(), config.Path, syscall.O_RDONLY, 0700)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening logger fifo: %q", config.Path)
	}
	p := &Input{
		started: false,
		outlet:  outlet,
		config:  config,
		cfg:     cfg,
		fifo:    f,
	}
	return p, nil
}

// LoadStates loads the states
// there is no state as we can only read until EOF
func (p *Input) LoadStates(states []file.State) error {
	return nil
}

// Run runs the input
func (p *Input) Run() {
	logp.Debug("dockerplugin", "Run dockerplugin input with path: %s", p.config.Path)
	if len(p.config.Path) == 0 {
		logp.Err("No path configured")
		return
	}
	forwarder := harvester.NewForwarder(p.outlet)
	p.consumeLog(forwarder)
}

// Stop stops the input and all its harvesters
func (p *Input) Stop() {
	p.outlet.Close()
}
func (p *Input) consumeLog(forwarder *harvester.Forwarder) {
	dec := protoio.NewUint32DelimitedReader(p.fifo, binary.BigEndian, 1e6)
	defer dec.Close()
	var buf logdriver.LogEntry
	for {
		if err := dec.ReadMsg(&buf); err != nil {
			if err == io.EOF {
				logp.Debug("dockerplugin", "End of stream %s", p.config.Path)
				p.fifo.Close()
				return
			}
			dec = protoio.NewUint32DelimitedReader(p.fifo, binary.BigEndian, 1e6)
		}
		forwarder.Send(createEvent(buf))
		buf.Reset()
	}
}
func createEvent(msg logdriver.LogEntry) *util.Data {
	data := util.NewData()
	data.Event = beat.Event{
		Timestamp: time.Unix(0, msg.TimeNano),
		Fields: common.MapStr{
			"message": string(msg.Line),
			"source":  string(msg.Source),
		},
	}
	return data
}

// Wait waits for the input to be completed. Not implemented.
func (p *Input) Wait() {}
