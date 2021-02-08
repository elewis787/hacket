package hacket

import "errors"

var (
	// ErrInvalidProtocol supplied protocol is invalid
	ErrInvalidProtocol = errors.New("invalid or unrecognizable protocol string")

	// ErrMissingPacketConn no connection was supplied or it is nil
	ErrMissingPacketConn = errors.New("mising packet connection or connection is nil")

	// ErrPacketServiceShutdown packet service is shutdown or in the the process of closing
	ErrPacketServiceShutdown = errors.New("packet service is shut down or in the process of shutting down")

	// ErrNilPacketHander cannot supply a nil packet handler
	ErrNilPacketHander = errors.New("nil packet handler")

	// ErrPacketHandlerNotFound missing or unable to find packet handler
	ErrPacketHandlerNotFound = errors.New("packet handler not found")

	// ErrPacketHandlerAlreadyExists cannot create packet handler because it already exists
	ErrPacketHandlerAlreadyExists = errors.New("packet handler already exists with supplied packet type")

	// ErrNilByteSlice Slice the byte slice passed was nil
	ErrNilByteSlice = errors.New("nil byte slice supplied")

	// ErrMaxMessageSize message is to large
	ErrMaxMessageSize = errors.New("message size excceed max packet size")

	//ErrByteWrite unable to write all of the bytes to the supplied packet
	ErrByteWrite = errors.New("failed to write all bytes to packet")

	//ErrNilConn is returned when trying to use the server with a nil connection
	ErrNilConn = errors.New("no packet connection")
)
