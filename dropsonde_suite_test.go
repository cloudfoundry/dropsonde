package dropsonde_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDropsonde(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dropsonde Suite")
}
