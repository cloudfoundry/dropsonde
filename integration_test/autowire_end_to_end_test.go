package integration_test

import (
	"bytes"
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry-incubator/dropsonde/autowire"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"
)

// these tests need to be invoked individually from an external script,
// since environment variables need to be set/unset before starting the tests
var _ = Describe("Autowire End-to-End", func() {
	Context("with DROPSONDE_ORIGIN missing", func() {
		var logWriter *bytes.Buffer
		BeforeEach(func() {
			logWriter = new(bytes.Buffer)
			log.SetOutput(logWriter)
		})

		Describe("init", func() {
			It("does not instrument http.DefaultTransport", func() {
				Expect(reflect.TypeOf(http.DefaultTransport).Elem().Name()).To(Equal("Transport"))
			})
		})

		Describe("InstrumentedHandler", func() {
			It("returns the given Handler with no changes and logs an error", func() {
				fake := FakeHandler{}

				Expect(autowire.InstrumentedHandler(fake)).To(Equal(fake))

				loggedText := string(logWriter.Bytes())

				expectedText := "Failed to instrument Handler; no emitter configured\n"
				Expect(loggedText).To(ContainSubstring(expectedText))
			})
		})

		Describe("InstrumentedRoundTripper", func() {
			It("returns the given RoundTripper with no changes and logs an error", func() {
				fake := FakeRoundTripper{}
				Expect(autowire.InstrumentedRoundTripper(fake)).To(Equal(fake))

				loggedText := string(logWriter.Bytes())

				expectedText := "Failed to instrument RoundTripper; no emitter configured\n"
				Expect(loggedText).To(ContainSubstring(expectedText))
			})
		})
	})

	Context("with DROPSONDE_ORIGIN set", func() {
		It("emits HTTP client/server events and heartbeats", func(done Done) {
			defer close(done)
			udpListener, err := net.ListenPacket("udp4", ":42420")
			Expect(err).ToNot(HaveOccurred())
			defer udpListener.Close()
			udpDataChan := make(chan []byte, 16)

			go func() {
				defer close(udpDataChan)
				for {
					buffer := make([]byte, 1024)
					n, _, _ := udpListener.ReadFrom(buffer)
					udpDataChan <- buffer[0:n]
				}
			}()

			httpListener, err := net.Listen("tcp", "localhost:0")
			Expect(err).ToNot(HaveOccurred())
			defer httpListener.Close()
			httpHandler := autowire.InstrumentedHandler(FakeHandler{})
			go http.Serve(httpListener, httpHandler)

			_, err = http.Get("http://" + httpListener.Addr().String())
			Expect(err).ToNot(HaveOccurred())

			expectedEvents := map[string]bool{
				"HttpStartClient": true,
				"HttpStopClient":  true,
				"HttpStartServer": true,
				"HttpStopServer":  true,
				"Heartbeat":       true,
			}

			for len(expectedEvents) > 0 {
				data := <-udpDataChan
				envelope := new(events.Envelope)
				err = proto.Unmarshal(data, envelope)
				Expect(err).ToNot(HaveOccurred())

				origin := os.Getenv("DROPSONDE_ORIGIN")
				Expect(envelope.GetOrigin()).To(Equal(origin))
				var eventId = envelope.GetEventType().String()

				switch envelope.GetEventType() {
				case events.Envelope_HttpStart:
					eventId += envelope.GetHttpStart().GetPeerType().String()
				case events.Envelope_HttpStop:
					eventId += envelope.GetHttpStop().GetPeerType().String()
				case events.Envelope_Heartbeat:
				default:
					panic("Unexpected message type")

				}

				delete(expectedEvents, eventId)
			}

		}, float64(emitter.HeartbeatInterval/time.Second)+1)
	})
})

type FakeHandler struct{}

func (fh FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "Hello")
}

type FakeRoundTripper struct{}

func (frt FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}
