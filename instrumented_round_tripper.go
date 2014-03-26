package dropsonde

import "net/http"

type instrumentedRoundTripper struct {
	rt http.RoundTripper
}

/*
Constructor for creating an InstrumentedRoundTripper which will delegate to the given RoundTripper
*/
func InstrumentedRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &instrumentedRoundTripper{rt: rt}
}

/*
Wraps the RoundTrip function of the given RoundTripper.
Will provide accounting metrics for the http.Request / http.Response life-cycle
Callers of RoundTrip are responsible for setting the ‘X-CF-RequestID’ field in the request header if they have one.
Callers are also responsible for setting the ‘X-CF-ApplicationID’ and ‘X-CF-InstanceIndex’ fields in the request header if they are known.
*/
func (irt *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// ...emit metrics here...
	return irt.rt.RoundTrip(req)
}
