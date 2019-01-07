package rpc

import (
	"context"
	"io"
)

type (
	// Encodable is an interface which must be implemented by user packets, which is sent over the wires.
	Encodable interface {
		// EncodedSize returns number of bytes the message needs to be encoded as a []byte
		EncodedSize() int
		// Encode
		Encode(writer io.Writer) error
	}

	// ClientCodec interface represents client connection on the client side.
	ClientCodec interface {
		io.Closer

		// WriteRequest encodes and writes the packet to the wire
		WriteRequest(reqId int32, funcId int16, msg Encodable) error
		// ReadResponse reads the server response header
		ReadResponse() (reqId int32, opErr error, bodySize int, err error)
		// ReadResponseBody reads the server response's body to the buffer provided. The len(body) must be the bodySize
		// returned by the ReadResponse for the response.
		ReadResponseBody(body []byte) error
	}

	// Client allows to make remote calls to the server
	Client interface {
		io.Closer
		BufCollector

		// Call makes a call to the server. Expects function Id, the message, which has to be sent. The Call will block the
		// calling go-routine until ctx is done, a response is received or an error happens related to the Call processing.
		// It returns response body, opErr contains the function execution error, which doesn't relate to the connection.
		// The err contains an error related to the call execution (connection problems etc.)
		Call(ctx context.Context, funcId int, msg Encodable) (respBody []byte, opErr error, err error)
	}

	// ServerCodec is an interface which represents the server connection on the server side
	ServerCodec interface {
		io.Closer

		// Id contains the Id for the server codec connection.
		Id() string

		ReadRequest() (reqId int32, funcId int16, bodySize int, err error)
		ReadRequestBody(body []byte) error
		WriteResponse(reqId int32, opErr error, msg Encodable) error
	}

	// Responder an interface provided to the ClientRequest callback function with a purpose to send response over there.
	Responder interface {
		// BufCollector interface is provided by the responder to be able to utilize request body, if needed
		BufCollector

		// SendResponse allows to send the response by the request id
		SendResponse(reqId int32, srvErr error, msg Encodable)
	}

	// OnClientReqFunc represents a server endpoint for handling a client call. The function will be selected by Server
	// by the funcId received in the request.
	OnClientReqFunc func(reqId int32, reqBody []byte, resp Responder)

	// Server supports server implementation for client requests handling.
	Server interface {
		io.Closer
		// Register adds a callback which will be called by the funcId
		Register(funcId int, cb OnClientReqFunc) error
		// Serve is called for sc - the server code which will be served until the codec is closed or the server is
		// shutdown
		Serve(sc ServerCodec) error
	}

	// BufCollector allows to collect byte buffers that are not going to be used anymore. The buffers could be
	// reused by Client or the Server later for processing requests and responses. Must be used with extra care
	BufCollector interface {
		// Collect stores the buf into the pool, to be reused for future needs
		Collect(buf []byte)
	}
)
