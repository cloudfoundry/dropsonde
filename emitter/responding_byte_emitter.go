package emitter

type RespondingByteEmitter interface {
	ByteEmitter
	RespondToPing()
}
