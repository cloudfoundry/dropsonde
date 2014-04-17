package emitter

import (
	"github.com/cloudfoundry-incubator/dropsonde/events"
)

type Emitter interface {
	Emit(events.Event) error
	Close()
}
