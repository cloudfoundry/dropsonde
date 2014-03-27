package events

import (
	"code.google.com/p/gogoprotobuf/proto"
	"encoding/binary"
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"net/http"
	"strconv"
	"time"
)

func NewUUID(id *uuid.UUID) *UUID {
	return &UUID{Low: proto.Uint64(binary.LittleEndian.Uint64(id[:8])), High: proto.Uint64(binary.LittleEndian.Uint64(id[8:]))}
}

func NewHttpStart(req *http.Request, peerType PeerType, requestId *uuid.UUID) *HttpStart {
	httpStart := &HttpStart{
		Timestamp:     proto.Int64(time.Now().UnixNano()),
		RequestId:     NewUUID(requestId),
		PeerType:      &peerType,
		Method:        HttpStart_Method(HttpStart_Method_value[req.Method]).Enum(),
		Uri:           proto.String(fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path)),
		RemoteAddress: proto.String(req.RemoteAddr),
		UserAgent:     proto.String(req.UserAgent()),
	}

	if applicationId, err := uuid.ParseHex(req.Header.Get("X-CF-ApplicationID")); err == nil {
		httpStart.ApplicationId = NewUUID(applicationId)
	}

	if instanceIndex, err := strconv.Atoi(req.Header.Get("X-CF-InstanceIndex")); err == nil {
		httpStart.InstanceIndex = proto.Int(instanceIndex)
	}

	return httpStart
}

func NewHttpStop(statusCode int, contentLength int64, peerType PeerType, requestId *uuid.UUID) *HttpStop {
	return &HttpStop{
		Timestamp:     proto.Int64(time.Now().UnixNano()),
		RequestId:     NewUUID(requestId),
		PeerType:      &peerType,
		StatusCode:    proto.Int(statusCode),
		ContentLength: proto.Int64(contentLength),
	}
}
