package signature_test

import (
	"github.com/cloudfoundry/dropsonde/signature"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SignatureVerifier", func() {
	var (
		inputChan         chan []byte
		outputChan        chan []byte
		runComplete       chan struct{}
		signatureVerifier signature.SignatureVerifier
	)

	BeforeEach(func() {
		inputChan = make(chan []byte, 10)
		outputChan = make(chan []byte, 10)
		runComplete = make(chan struct{})
		signatureVerifier = signature.NewSignatureVerifier(loggertesthelper.Logger())

		go func() {
			signatureVerifier.Run(inputChan, outputChan)
			close(runComplete)
		}()
	})

	AfterEach(func() {
		close(inputChan)
		Eventually(runComplete).Should(BeClosed())
	})

	It("removes first 32 bytes", func() {
		message := make([]byte, 33)
		message[32] = 123
		inputChan <- message
		num := <-outputChan
		Expect(num).To(HaveLen(1))
		Expect(num[0]).To(Equal(byte(123)))
	})

	It("discards messages less than 32 bytes long", func() {
		message := make([]byte, 1)
		inputChan <- message
		Consistently(outputChan).ShouldNot(Receive())
	})
})
