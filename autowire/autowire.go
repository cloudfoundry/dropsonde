package autowire

import (
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/runtime_stats"
	"log"
	"net/http"
	"os"
	"time"
)

var autowiredEmitter emitter.EventEmitter

const runtimeStatsInterval = 10 * time.Second

var destination string

const defaultDestination = "localhost:42420"

func init() {
	Initialize()
}

func InstrumentedHandler(handler http.Handler) http.Handler {
	if autowiredEmitter == nil {
		return handler
	}

	return dropsonde.InstrumentedHandler(handler, autowiredEmitter)
}

func InstrumentedRoundTripper(roundTripper http.RoundTripper) http.RoundTripper {
	if autowiredEmitter == nil {
		return roundTripper
	}

	return dropsonde.InstrumentedRoundTripper(roundTripper, autowiredEmitter)
}

func Destination() string {
	return destination
}

func Initialize() {
	http.DefaultTransport = &http.Transport{Proxy: http.ProxyFromEnvironment}
	autowiredEmitter = nil

	origin := os.Getenv("DROPSONDE_ORIGIN")
	if len(origin) == 0 {
		log.Println("Failed to auto-initialize dropsonde: DROPSONDE_ORIGIN environment variable not set")
		return
	}

	destination = os.Getenv("DROPSONDE_DESTINATION")
	if len(destination) == 0 {
		log.Println("DROPSONDE_DESTINATION not set. Using " + defaultDestination)
		destination = defaultDestination
	}

	udpEmitter, err := emitter.NewUdpEmitter(destination)
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

	go runtime_stats.NewRuntimeStats(autowiredEmitter, runtimeStatsInterval).Run(nil)

	http.DefaultTransport = InstrumentedRoundTripper(http.DefaultTransport)
}

func AutowiredEmitter() emitter.EventEmitter {
	return autowiredEmitter
}
