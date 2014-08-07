package dropsonde_unmarshaller_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/dropsonde/dropsonde_unmarshaller"
	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DropsondeUnmarshaller", func() {
	var (
		inputChan    chan []byte
		outputChan   chan *events.Envelope
		runComplete  chan struct{}
		unmarshaller dropsonde_unmarshaller.DropsondeUnmarshaller
	)

	BeforeEach(func() {
		inputChan = make(chan []byte, 10)
		outputChan = make(chan *events.Envelope, 10)
		runComplete = make(chan struct{})
		unmarshaller = dropsonde_unmarshaller.NewDropsondeUnmarshaller(loggertesthelper.Logger())

		go func() {
			unmarshaller.Run(inputChan, outputChan)
			close(runComplete)
		}()
	})

	AfterEach(func() {
		close(inputChan)
		Eventually(runComplete).Should(BeClosed())
	})

	It("unmarshals bytes into envelopes", func() {
		envelope := &events.Envelope{
			Origin:    proto.String("fake-origin-3"),
			EventType: events.Envelope_Heartbeat.Enum(),
			Heartbeat: factories.NewHeartbeat(1, 2, 3),
		}
		message, _ := proto.Marshal(envelope)

		inputChan <- message
		outputEnvelope := <-outputChan
		Expect(outputEnvelope).To(Equal(envelope))
	})

	Context("metrics", func() {
		var metricValue = func(name string) interface{} {
			for _, metric := range unmarshaller.Emit().Metrics {
				if metric.Name == name {
					return metric.Value
				}
			}
			return nil
		}

		var eventuallyExpectMetric = func(name string, value uint64) {
			Eventually(func() interface{} {
				return metricValue(name)
			}).Should(Equal(value))
		}

		It("emits the correct metrics context", func() {
			Expect(unmarshaller.Emit().Name).To(Equal("dropsondeUnmarshaller"))
		})

		It("emits a heartbeat counter", func() {
			envelope := &events.Envelope{
				Origin:    proto.String("fake-origin-3"),
				EventType: events.Envelope_Heartbeat.Enum(),
				Heartbeat: factories.NewHeartbeat(1, 2, 3),
			}
			message, _ := proto.Marshal(envelope)

			inputChan <- message
			eventuallyExpectMetric("heartbeatReceived", 1)
		})

		It("emits an unmarshal error counter", func() {
			inputChan <- []byte{1, 2, 3}
			eventuallyExpectMetric("unmarshalErrors", 1)
		})
	})
})
