package autowire_test

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/dropsonde/autowire"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/dropsonde/control"
	"github.com/cloudfoundry/dropsonde/factories"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Autowire", func() {
	var oldDestination string
	var oldOrigin string

	BeforeEach(func() {
		oldDestination = os.Getenv("DROPSONDE_DESTINATION")
		oldOrigin = os.Getenv("DROPSONDE_ORIGIN")
	})

	AfterEach(func() {
		os.Setenv("DROPSONDE_DESTINATION", oldDestination)
		os.Setenv("DROPSONDE_ORIGIN", oldOrigin)
	})

	Describe("Initialize", func() {
		Context("with a non-nil emitter", func() {
			It("instruments the HTTP default transport", func() {
				autowire.Initialize(emitter.NewEventEmitter(nil, ""))
				Expect(reflect.TypeOf(http.DefaultTransport).Elem().Name()).ToNot(Equal("Transport"))
			})
		})

		Context("with a nil-emitter", func() {
			It("resets the HTTP default transport to not be instrumented", func() {
				autowire.Initialize(nil)
				Expect(reflect.TypeOf(http.DefaultTransport).Elem().Name()).To(Equal("Transport"))
			})
		})
	})

	Describe("CreateDefaultEmitter", func() {
		Context("with DROPSONDE_ORIGIN set", func() {
			BeforeEach(func() {
				os.Setenv("DROPSONDE_ORIGIN", "anything")
			})

			Context("with DROPSONDE_DESTINATION missing", func() {
				It("defaults to localhost", func() {
					os.Setenv("DROPSONDE_DESTINATION", "")
					_, destination := autowire.CreateDefaultEmitter()

					Expect(destination).To(Equal("localhost:3457"))
				})
			})

			Context("with DROPSONDE_DESTINATION set", func() {
				It("uses the configured destination", func() {
					os.Setenv("DROPSONDE_DESTINATION", "test")
					_, destination := autowire.CreateDefaultEmitter()

					Expect(destination).To(Equal("test"))
				})
			})

			It("responds to heartbeat requests with heartbeats", func() {
				os.Setenv("DROPSONDE_DESTINATION", "localhost:1235")

				messages := make(chan []byte, 100)
				readyChan := make(chan struct{})

				go respondWithHeartbeatRequest(1235, messages, readyChan)
				<-readyChan

				emitter, _ := autowire.CreateDefaultEmitter()

				err := emitter.Emit(&events.CounterEvent{Name: proto.String("name"), Delta: proto.Uint64(1)})
				Expect(err).NotTo(HaveOccurred())

				Eventually(messages, 5).Should(Receive())
			})
		})

		Context("with DROPSONDE_ORIGIN missing", func() {
			It("returns a nil-emitter", func() {
				os.Setenv("DROPSONDE_ORIGIN", "")
				emitter, _ := autowire.CreateDefaultEmitter()
				Expect(emitter).To(BeNil())
			})
		})
	})
})

type FakeHandler struct{}

func (fh FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {}

type FakeRoundTripper struct{}

func (frt FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func respondWithHeartbeatRequest(port int, messages chan []byte, readyChan chan struct{}) {
	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	close(readyChan)
	n, addr, _ := conn.ReadFrom(buf)

	conn.WriteTo(newMarshalledHeartbeatRequest(), addr)
	n, addr, _ = conn.ReadFrom(buf)

	messages <- buf[:n]
	conn.Close()
}

func newMarshalledHeartbeatRequest() []byte {
	id, _ := uuid.NewV4()

	heartbeatRequest := &control.ControlMessage{
		Origin:      proto.String("test"),
		Identifier:  factories.NewControlUUID(id),
		Timestamp:   proto.Int64(time.Now().UnixNano()),
		ControlType: control.ControlMessage_HeartbeatRequest.Enum(),
	}

	bytes, err := proto.Marshal(heartbeatRequest)
	if err != nil {
		panic(err.Error())
	}
	return bytes
}
