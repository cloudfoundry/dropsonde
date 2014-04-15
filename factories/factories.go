package factories

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"encoding/binary"
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"net/http"
	"strconv"
	"time"
)

func NewUUID(id *uuid.UUID) *events.UUID {
	return &events.UUID{Low: proto.Uint64(binary.LittleEndian.Uint64(id[:8])), High: proto.Uint64(binary.LittleEndian.Uint64(id[8:]))}
}

func NewHttpStart(req *http.Request, peerType events.PeerType, requestId *uuid.UUID) *events.HttpStart {
	httpStart := &events.HttpStart{
		Timestamp:     proto.Int64(time.Now().UnixNano()),
		RequestId:     NewUUID(requestId),
		PeerType:      &peerType,
		Method:        events.HttpStart_Method(events.HttpStart_Method_value[req.Method]).Enum(),
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

func NewHttpStop(statusCode int, contentLength int64, peerType events.PeerType, requestId *uuid.UUID) *events.HttpStop {
	return &events.HttpStop{
		Timestamp:     proto.Int64(time.Now().UnixNano()),
		RequestId:     NewUUID(requestId),
		PeerType:      &peerType,
		StatusCode:    proto.Int(statusCode),
		ContentLength: proto.Int64(contentLength),
	}
}

func NewHeartbeat(sentCount, receivedCount, errorCount uint64) *events.Heartbeat {
	return &events.Heartbeat{
		SentCount:     proto.Uint64(sentCount),
		ReceivedCount: proto.Uint64(receivedCount),
		ErrorCount:    proto.Uint64(errorCount),
	}
}
