package dropsonde_test

import (
	"net/http"
	"reflect"

	"github.com/cloudfoundry/dropsonde"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Autowire", func() {

	Describe("Initialize", func() {
		It("resets the HTTP default transport to be instrumented", func() {
			dropsonde.InitializeWithEmitter(&dropsonde.NullEventEmitter{})
			Expect(reflect.TypeOf(http.DefaultTransport).Elem().Name()).To(Equal("instrumentedCancelableRoundTripper"))
		})
	})

	Describe("CreateDefaultEmitter", func() {
		DescribeTable("returns a NullEventEmitter",
			func(origin, deployment, job, index string) {
				err := dropsonde.Initialize("localhost:2343", origin, deployment, job, index)
				Expect(err).To(HaveOccurred())

				emitter := dropsonde.AutowiredEmitter()
				Expect(emitter).ToNot(BeNil())
				nullEmitter := &dropsonde.NullEventEmitter{}
				Expect(emitter).To(BeAssignableToTypeOf(nullEmitter))
			},
			Entry("empty origin", "", "deployment", "job", "index"),
			Entry("empty deployment", "origin", "", "job", "index"),
			Entry("empty job", "origin", "deployment", "", "index"),
			Entry("empty index", "origin", "deployment", "job", ""),
		)
	})
})

type FakeHandler struct{}

func (fh FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {}

type FakeRoundTripper struct{}

func (frt FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}
