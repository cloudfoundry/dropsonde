package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"sync/atomic"
)

type SignatureVerifier interface {
	instrumentation.Instrumentable
	Run(inputChan <-chan []byte, outputChan chan<- []byte)
}

func NewSignatureVerifier(logger *gosteno.Logger, sharedSecret string) SignatureVerifier {
	return &signatureVerifier{
		logger:       logger,
		sharedSecret: sharedSecret,
	}
}

type signatureVerifier struct {
	logger                     *gosteno.Logger
	sharedSecret               string
	missingSignatureErrorCount uint64
	invalidSignatureErrorCount uint64
	validSignatureCount        uint64
}

func (v *signatureVerifier) Run(inputChan <-chan []byte, outputChan chan<- []byte) {
	for signedMessage := range inputChan {
		if len(signedMessage) < 32 {
			v.logger.Warnf("signatureVerifier: missing signature for message %v", signedMessage)
			incrementCount(&v.missingSignatureErrorCount)
			continue
		}

		signature, message := signedMessage[:32], signedMessage[32:]
		if v.verifyMessage(message, signature) {
			outputChan <- message
			incrementCount(&v.validSignatureCount)
			v.logger.Debugf("signatureVerifier: valid signature %v for message %v", signature, message)
		} else {
			v.logger.Warnf("signatureVerifier: invalid signature %v for message %v", signature, message)
			incrementCount(&v.invalidSignatureErrorCount)
		}
	}
}

func (v *signatureVerifier) verifyMessage(message, signature []byte) bool {
	expectedMAC := generateSignature(message, []byte(v.sharedSecret))
	return hmac.Equal(signature, expectedMAC)
}

func (v *signatureVerifier) metrics() []instrumentation.Metric {
	return []instrumentation.Metric{
		instrumentation.Metric{Name: "missingSignatureErrors", Value: atomic.LoadUint64(&v.missingSignatureErrorCount)},
		instrumentation.Metric{Name: "invalidSignatureErrors", Value: atomic.LoadUint64(&v.invalidSignatureErrorCount)},
		instrumentation.Metric{Name: "validSignatures", Value: atomic.LoadUint64(&v.validSignatureCount)},
	}
}

func (v *signatureVerifier) Emit() instrumentation.Context {
	return instrumentation.Context{
		Name:    "signatureVerifier",
		Metrics: v.metrics(),
	}
}

func SignMessage(message, secret []byte) []byte {
	signature := generateSignature(message, secret)
	return append(signature, message...)
}

func generateSignature(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}

func incrementCount(count *uint64) {
	atomic.AddUint64(count, 1)
}
