package signature

import (
	"github.com/cloudfoundry/gosteno"
)

type SignatureVerifier interface {
	Run(inputChan <-chan []byte, outputChan chan<- []byte)
}

func NewSignatureVerifier(logger *gosteno.Logger) SignatureVerifier {
	return &signatureVerifier{
		logger: logger,
	}
}

type signatureVerifier struct {
	logger *gosteno.Logger
}

func (u *signatureVerifier) Run(inputChan <-chan []byte, outputChan chan<- []byte) {
	for message := range inputChan {
		if len(message) >= 32 {
			// TODO: log errors (and maybe publish metric) for messages with missing signature
			outputChan <- message[32:]
		}
	}
}
