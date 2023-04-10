package signature_test

import (
	"crypto/hmac"
	"crypto/sha256"

	. "github.com/apoydence/eachers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/dropsonde/metrics"
	"github.com/cloudfoundry/dropsonde/signature"
)

var _ = Describe("Verifier", func() {
	var (
		inputChan   chan []byte
		outputChan  chan []byte
		runComplete chan struct{}

		signatureVerifier *signature.Verifier

		mockBatcher *mockMetricBatcher
	)

	BeforeEach(func() {
		inputChan = make(chan []byte, 10)
		outputChan = make(chan []byte, 10)
		runComplete = make(chan struct{})

		signatureVerifier = signature.NewVerifier("valid-secret")

		mockBatcher = newMockMetricBatcher()
		metrics.Initialize(nil, mockBatcher)

		go func() {
			signatureVerifier.Run(inputChan, outputChan)
			close(runComplete)
		}()
	})

	AfterEach(func() {
		close(inputChan)
		Eventually(runComplete).Should(BeClosed())
	})

	It("discards messages less than 32 bytes long", func() {
		message := make([]byte, 1)
		inputChan <- message
		Consistently(outputChan).ShouldNot(Receive())
	})

	It("discards messages when verification fails", func() {
		message := make([]byte, 33)

		inputChan <- message
		Consistently(outputChan).ShouldNot(Receive())
	})

	It("passes through messages with valid signature", func() {
		message := []byte{1, 2, 3}
		mac := hmac.New(sha256.New, []byte("valid-secret"))
		mac.Write(message)
		signature := mac.Sum(nil)

		signedMessage := append(signature, message...)

		inputChan <- signedMessage
		outputMessage := <-outputChan
		Expect(outputMessage).To(Equal(message))
	})

	Context("metrics", func() {
		It("emits an valid signature counter", func() {
			message := []byte{1, 2, 3}
			mac := hmac.New(sha256.New, []byte("valid-secret"))
			mac.Write(message)
			signature := mac.Sum(nil)

			signedMessage := append(signature, message...)
			inputChan <- signedMessage

			Eventually(mockBatcher.BatchIncrementCounterInput).Should(BeCalled(
				With("signatureVerifier.validSignatures"),
			))
		})
	})
})
