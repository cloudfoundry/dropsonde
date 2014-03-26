package dropsonde

import (
	"net/http"
	uuid "github.com/nu7hatch/gouuid"
	"encoding/binary"
)

type instrumentedHandler struct {
	h http.Handler
}

/*
Constructor for creating an Instrumented Handler which will delegate to the given http.Handler.
*/
func InstrumentedHandler(h http.Handler) http.Handler {
	return &instrumentedHandler{h}
}

/*
Wraps the given http.Handler ServerHTTP function
Will provide accounting metrics for the http.Request / http.Response life-cycle
*/
func (ih *instrumentedHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// ...emit metrics here...

	requestId := req.Header.Get("X-CF-RequestID")
	if requestId == "" {
		id, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}
		requestId = id.String()
		lb, hb := split(id)
		println(hb, lb)
		req.Header.Set("X-CF-RequestID", requestId)
	}
	rw.Header().Set("X-CF-RequestID", requestId)

	// create http start event
	// send start event via emitter?

	ih.h.ServeHTTP(rw, req)
	// send stop event
}


func split(id *uuid.UUID) (uint64, uint64) {
	return binary.LittleEndian.Uint64(id[:8]), binary.LittleEndian.Uint64(id[8:])
}
