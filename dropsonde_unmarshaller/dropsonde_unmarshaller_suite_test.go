package dropsonde_unmarshaller_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/loggregatorlib/cfcomponent"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"testing"
)

func TestUnmarshaller(t *testing.T) {
	cfcomponent.Logger = loggertesthelper.Logger()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dropsonde Unmarshaller Suite")
}
