package metricbatcher_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMetricBatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetricBatcher Suite")
}
