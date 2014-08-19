// Package autowire provides sensible defaults for using dropsonde.
//
// The default HTTP transport is instrumented, as well as some basic stats about
// the Go runtime. Additionally, the default emitter is itself instrumented to
// periodically send "heartbeat" messages containing counts of received and sent
// events. The default emitter sends events over UDP.
//
// Use
//
// Set the DROPSONDE_ORIGIN and DROPSONDE_DESTINATION environment variables.
// (See Initialize below for details.) Anonymously import autowire:
//
//		import (
// 			_ "github.com/cloudfoundry/dropsonde/autowire"
// 		)
//
// The package self-initializes and automatically adds instrumentation where it
// can.
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

const defaultDestination = "localhost:3457"

func init() {
	Initialize()
}

// InstrumentedHandler returns a Handler pre-configured to emit to autowire's
// Emitter (if one has been set), or the original Handler (if one has been set),
// or the original handler (if autowire has not been initialized).
func InstrumentedHandler(handler http.Handler) http.Handler {
	if autowiredEmitter == nil {
		return handler
	}

	return dropsonde.InstrumentedHandler(handler, autowiredEmitter)
}

// InstrumentedRoundTripper returns a RoundTripper pre-configured to emit to
// autowire's Emitter (if one has been set), or the original RoundTripper (if
// autowire has not been initialized).
func InstrumentedRoundTripper(roundTripper http.RoundTripper) http.RoundTripper {
	if autowiredEmitter == nil {
		return roundTripper
	}

	return dropsonde.InstrumentedRoundTripper(roundTripper, autowiredEmitter)
}

func Destination() string {
	return destination
}

// Initialize creates default emitters and instruments the default HTTP
// transport.
//
// The DROPSONDE_ORIGIN environment variable is required and specifies the
// source name for all metrics emitted by this process.
//
// The DROPSONDE_DESTINATION environment variable sets the host and port to
// which metrics are sent. It is optional, and defaults to localhost:42420.
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
