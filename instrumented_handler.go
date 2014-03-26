package dropsonde

import (
	"net/http"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
	"code.google.com/p/gogoprotobuf/proto"
	"time"
	"fmt"
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
	requestId, err := uuid.ParseHex(req.Header.Get("X-CF-RequestID"))
	if err != nil {
		requestId, err = uuid.NewV4()
		if err != nil {
			panic(err)
		}
		req.Header.Set("X-CF-RequestID", requestId.String())
	}
	rw.Header().Set("X-CF-RequestID", requestId.String())

	emitter.Emit(CreateHttpStart(req, requestId))

	instrumentedWriter := &instrumentedResponseWriter{writer: rw, statusCode: 200}
	ih.h.ServeHTTP(instrumentedWriter, req)

	emitter.Emit(CreateHttpStop(instrumentedWriter, requestId))
}

type instrumentedResponseWriter struct {
	writer http.ResponseWriter
	contentLength, statusCode int
}

func (irw *instrumentedResponseWriter) Header() http.Header {
	return irw.writer.Header()
}

func (irw *instrumentedResponseWriter) Write(data []byte) (int, error) {
	writeCount, err :=  irw.writer.Write(data)
	irw.contentLength += writeCount
	return writeCount, err
}

func (irw *instrumentedResponseWriter) WriteHeader(statusCode int) {
	irw.statusCode = statusCode
	irw.writer.WriteHeader(statusCode)
}

func CreateHttpStart(req *http.Request, requestId *uuid.UUID) *events.HttpStart {
	return &events.HttpStart{
		Timestamp: proto.Int64(time.Now().UnixNano()),
		RequestId: events.NewUUID(requestId),
		PeerType: events.PeerType_Server.Enum(),
		Method: events.HttpStart_Method(events.HttpStart_Method_value[req.Method]).Enum(),
		Uri: proto.String(fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path)),
		RemoteAddress: proto.String(req.RemoteAddr),
		UserAgent: proto.String(req.UserAgent()),
	}
}

func CreateHttpStop(irw *instrumentedResponseWriter, requestId *uuid.UUID) *events.HttpStop {
	return &events.HttpStop{
		Timestamp: proto.Int64(time.Now().UnixNano()),
		RequestId: events.NewUUID(requestId),
		PeerType: events.PeerType_Server.Enum(),
		StatusCode: proto.Int(irw.statusCode),
		ContentLength: proto.Int(irw.contentLength),
	}
}
