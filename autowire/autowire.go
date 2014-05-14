package autowire

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde-common/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/udp_emitter"
	"log"
	"net/http"
	"os"
)

var autowiredEmitter emitter.EventEmitter

const autowiredEmitterRemoteAddr = "localhost:42420"

func init() {
	origin := os.Getenv("DROPSONDE_ORIGIN")
	if len(origin) == 0 {
		log.Printf("Failed to auto-initialize dropsonde: DROPSONDE_ORIGIN environment variable not set\n")
		return
	}

	udpEmitter, err := udp_emitter.NewUdpEmitter(autowiredEmitterRemoteAddr)
	if err != nil {
		log.Printf("Failed to auto-initialize dropsonde: %v\n", err)
		return
	}

	hbEmitter, err := emitter.NewHeartbeatEmitter(udpEmitter, origin)
	if err != nil {
		log.Printf("Failed to auto-initialize dropsonde: %v\n", err)
		return
	}

	autowiredEmitter = emitter.NewEventEmitter(hbEmitter, origin)

	http.DefaultTransport = InstrumentedRoundTripper(http.DefaultTransport)
}

func InstrumentedHandler(handler http.Handler) http.Handler {
	if autowiredEmitter == nil {
		log.Printf("Failed to instrument Handler; no emitter configured\n")
		return handler
	}

	return dropsonde.InstrumentedHandler(handler, autowiredEmitter)
}

func InstrumentedRoundTripper(roundTripper http.RoundTripper) http.RoundTripper {
	if autowiredEmitter == nil {
		log.Printf("Failed to instrument RoundTripper; no emitter configured\n")
		return roundTripper
	}

	return dropsonde.InstrumentedRoundTripper(roundTripper, autowiredEmitter)
}
