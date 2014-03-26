package hunter

import (
	"io/ioutil"
	"log"
	"net/http"
)

type Handler struct {
	Client *http.Client
}

func NewHandler(client *http.Client) *Handler {
	return &Handler{
		Client: client,
	}
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
