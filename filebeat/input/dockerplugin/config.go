package dockerplugin

var defaultConfig = config{}

type config struct {
	Path string `config:"path"`
}
