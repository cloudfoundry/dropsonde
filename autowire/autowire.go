package autowire

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"log"
	"net/http"
	"os"
)

var autowiredEmitter emitter.Emitter

const autowiredEmitterRemoteAddr = "localhost:42420"

func init() {
	origin := os.Getenv("DROPSONDE_ORIGIN")
	if len(origin) == 0 {
		log.Printf("Failed to auto-initialize dropsonde: DROPSONDE_ORIGIN environment variable not set\n")
		return
	}

	udpEmitter, err := emitter.NewUdpEmitter(autowiredEmitterRemoteAddr, origin)
	if err != nil {
		log.Printf("Failed to auto-initialize dropsonde: %v\n", err)
		return
	}

	autowiredEmitter, err = emitter.NewHeartbeatEmitter(udpEmitter)
	if err != nil {
		log.Printf("Failed to auto-initialize dropsonde: %v\n", err)
		return
	}

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
