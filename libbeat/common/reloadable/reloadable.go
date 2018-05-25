package reloadable

import (
	"github.com/elastic/beats/libbeat/common"
)

// Reloadable objects can do live updates to their config
type Reloadable interface {
	Reload(*common.Config) error
}

// List of objects that can do live updates to their config
type List interface {
	Reload([]*common.Config) error
}
