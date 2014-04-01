package dropsonde

import (
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	uuid "github.com/nu7hatch/gouuid"
	"net/http"
)

type instrumentedHandler struct {
	h      http.Handler
	origin events.Origin
}

/*
Helper for creating an Instrumented Handler which will delegate to the given http.Handler.
*/
func InstrumentedHandler(h http.Handler, jobName string, jobIndex int32) http.Handler {
	origin := events.Origin{JobName: &jobName, JobInstanceId: &jobIndex}
	return &instrumentedHandler{h, origin}
}

/*
Wraps the given http.Handler ServerHTTP function
Will provide accounting metrics for the http.Request / http.Response life-cycle
*/
func (ih *instrumentedHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestId, err := uuid.ParseHex(req.Header.Get("X-CF-RequestID"))
	if err != nil {
		requestId, err = uuid.NewV4()
		if err != nil {
			panic(err)
		}
		req.Header.Set("X-CF-RequestID", requestId.String())
	}
	rw.Header().Set("X-CF-RequestID", requestId.String())

	startEvent := events.NewHttpStart(req, events.PeerType_Server, requestId)
	emitter.Emit(startEvent, ih.origin)

	instrumentedWriter := &instrumentedResponseWriter{writer: rw, statusCode: 200}
	ih.h.ServeHTTP(instrumentedWriter, req)

	stopEvent := events.NewHttpStop(instrumentedWriter.statusCode, instrumentedWriter.contentLength,
		events.PeerType_Server, requestId)

	emitter.Emit(stopEvent, ih.origin)
}

type instrumentedResponseWriter struct {
	writer        http.ResponseWriter
	contentLength int64
	statusCode    int
}

func (irw *instrumentedResponseWriter) Header() http.Header {
	return irw.writer.Header()
}

func (irw *instrumentedResponseWriter) Write(data []byte) (int, error) {
	writeCount, err := irw.writer.Write(data)
	irw.contentLength += int64(writeCount)
	return writeCount, err
}

func (irw *instrumentedResponseWriter) WriteHeader(statusCode int) {
	irw.statusCode = statusCode
	irw.writer.WriteHeader(statusCode)
}
