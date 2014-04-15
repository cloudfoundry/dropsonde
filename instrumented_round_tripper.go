package dropsonde

import (
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	uuid "github.com/nu7hatch/gouuid"
	"net/http"
)

type instrumentedRoundTripper struct {
	rt http.RoundTripper
}

/*
Helper for creating an InstrumentedRoundTripper which will delegate to the given RoundTripper
*/
func InstrumentedRoundTripper(rt http.RoundTripper) (http.RoundTripper, error) {
	return &instrumentedRoundTripper{rt}, nil
}

/*
Wraps the RoundTrip function of the given RoundTripper.
Will provide accounting metrics for the http.Request / http.Response life-cycle
Callers of RoundTrip are responsible for setting the ‘X-CF-RequestID’ field in the request header if they have one.
Callers are also responsible for setting the ‘X-CF-ApplicationID’ and ‘X-CF-InstanceIndex’ fields in the request header if they are known.
*/
func (irt *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	requestId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	httpStart := events.NewHttpStart(req, events.PeerType_Client, requestId)

	parentRequestId, err := uuid.ParseHex(req.Header.Get("X-CF-RequestID"))
	if err == nil {
		httpStart.ParentRequestId = events.NewUUID(parentRequestId)
	}

	req.Header.Set("X-CF-RequestID", requestId.String())

	emitter.Emit(httpStart)

	resp, err := irt.rt.RoundTrip(req)

	var httpStop *events.HttpStop
	if err != nil {
		httpStop = events.NewHttpStop(0, 0, events.PeerType_Client, requestId)
	} else {
		httpStop = events.NewHttpStop(resp.StatusCode, resp.ContentLength, events.PeerType_Client, requestId)
	}

	emitter.Emit(httpStop)

	return resp, err
}
