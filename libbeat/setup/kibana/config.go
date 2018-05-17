package kibana

import (
	"time"

	"github.com/elastic/beats/libbeat/outputs"
)

// Config used to connect to Kibana
type Config struct {
	Protocol string             `config:"protocol"`
	Host     string             `config:"host"`
	Path     string             `config:"path"`
	Username string             `config:"username"`
	Password string             `config:"password"`
	TLS      *outputs.TLSConfig `config:"ssl"`
	Timeout  time.Duration      `config:"timeout"`
}

var (
	defaultKibanaConfig = Config{
		Protocol: "http",
		Host:     "localhost:5601",
		Path:     "",
		Username: "",
		Password: "",
		Timeout:  90 * time.Second,
		TLS:      nil,
	}
)
