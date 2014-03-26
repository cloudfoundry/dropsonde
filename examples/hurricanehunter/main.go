package main

import (
	"github.com/cloudfoundry/dropsonde/examples/hurricanehunter/hunter"
	"log"
	"net/http"
)

func main() {
	log.Print("Launching HurricaneHunter … testing dropsondes")

	hunter := hunter.NewHandler(http.DefaultClient)
	log.Fatal(http.ListenAndServe(":8080", hunter))
}
