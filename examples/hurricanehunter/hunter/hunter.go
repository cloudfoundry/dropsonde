package hunter

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"io/ioutil"
	"log"
	"net/http"
)

func init() {
	http.DefaultTransport = dropsonde.InstrumentedRoundTripper(http.DefaultTransport)
}

type Handler struct {
	Client http.Client
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Panic("ParseForm error: ", err)
	}

	url := r.FormValue("url")
	log.Println(url)

	resp, err := h.Client.Get(url)
	if err != nil {
		log.Panic("Get error: ", err)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic("Read error: ", err)
	}

	w.Write(bytes)
}
