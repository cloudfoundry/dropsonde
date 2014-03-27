package main

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde/examples/hurricanehunter/hunter"
	"log"
	"net/http"
)

func main() {
	log.Print("Launching HurricaneHunter on port 8080 … testing dropsondes")

	handler := hunter.NewHandler(&http.Client{Transport: dropsonde.InstrumentedRoundTripper(http.DefaultTransport)})
	instrumentedHunter := dropsonde.InstrumentedHandler(handler)
	log.Fatal(http.ListenAndServe(":8080", instrumentedHunter))
}
