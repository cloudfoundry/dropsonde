package runtime_stats_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRuntimeStats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RuntimeStats Suite")
}
