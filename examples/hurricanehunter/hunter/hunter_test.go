package hunter_test

import (
	"bytes"
	"fmt"
	"github.com/cloudfoundry/dropsonde/examples/hurricanehunter/hunter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
)

var _ = Describe("Hunter", func() {
	var handler *hunter.Handler
	var recorder *httptest.ResponseRecorder
	var server *httptest.Server
	BeforeEach(func() {
		handler = hunter.NewHandler(http.DefaultClient)
		recorder = httptest.NewRecorder()
		server = httptest.NewServer(new(testHandler))
	})

	It("Echoes the response from GETting a URL", func() {
		data := url.Values{}
		data.Set("url", server.URL)

		req, err := http.NewRequest("POST", "http://localhost:8081", bytes.NewBufferString(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		if err != nil {
			panic(err)
		}

		handler.ServeHTTP(recorder, req)

		Expect(recorder.Body.String()).To(Equal("Hello"))
	})
})

type testHandler struct{}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}
