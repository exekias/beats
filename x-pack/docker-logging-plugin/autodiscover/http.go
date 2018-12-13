package autodiscover

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/docker/docker/daemon/logger"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/bus"
)

func init() {
	containers = containersMap{
		running: make(map[string]*logger.Info),
	}
}

var containers containersMap

type containersMap struct {
	sync.Mutex
	running map[string]*logger.Info
}
type StartLoggingRequest struct {
	File string
	Info logger.Info
}
type StopLoggingRequest struct {
	File string
}
type CapabilitiesResponse struct {
	Err string
	Cap logger.Capability
}
type ReadLogsRequest struct {
	Info   logger.Info
	Config logger.ReadConfig
}
type response struct {
	Err string
}

func initHandlers(p *Provider) {
	p.handler.HandleFunc("/LogDriver.StartLogging", func(w http.ResponseWriter, r *http.Request) {
		var req StartLoggingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Info.ContainerID == "" {
			respond(errors.New("must provide container id in log context"), w)
			return
		}
		containers.Lock()
		containers.running[req.File] = &req.Info
		containers.Unlock()
		p.publish(genEvent(req.File, &req.Info, "start"))
		respond(nil, w)
	})
	p.handler.HandleFunc("/LogDriver.StopLogging", func(w http.ResponseWriter, r *http.Request) {
		var req StopLoggingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		containers.Lock()
		info, ok := containers.running[req.File]
		delete(containers.running, req.File)
		containers.Unlock()
		if !ok {
			respond(errors.New("unknown container"), w)
			return
		}
		p.publish(genEvent(req.File, info, "stop"))
		respond(nil, w)
	})
	p.handler.HandleFunc("/LogDriver.Capabilities", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&CapabilitiesResponse{
			Cap: logger.Capability{ReadLogs: false},
		})
	})
}
func genEvent(file string, info *logger.Info, flag string) bus.Event {
	labels := common.MapStr{}
	for k, v := range info.ContainerLabels {
		labels[k] = v
	}
	meta := common.MapStr{
		"container": common.MapStr{
			"id":     info.ContainerID,
			"name":   info.ContainerName,
			"image":  info.ContainerImageName,
			"labels": labels,
		},
	}
	return bus.Event{
		flag:     true,
		"id":     info.ContainerID,
		"path":   file,
		"docker": meta,
		"meta": common.MapStr{
			"docker": meta,
		},
	}
}
func respond(err error, w http.ResponseWriter) {
	var res response
	if err != nil {
		res.Err = err.Error()
	}
	json.NewEncoder(w).Encode(&res)
}
