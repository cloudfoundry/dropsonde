package events

import (
	"encoding/binary"
	uuid "github.com/nu7hatch/gouuid"
	"code.google.com/p/gogoprotobuf/proto"
)

func NewUUID(id *uuid.UUID) *UUID {
	return &UUID{Low: proto.Uint64(binary.LittleEndian.Uint64(id[:8])), High: proto.Uint64(binary.LittleEndian.Uint64(id[8:]))}
}
