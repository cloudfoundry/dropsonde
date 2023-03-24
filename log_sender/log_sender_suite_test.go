package log_sender_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMetricSender(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogSender Suite")
}
