package ezconn

import (
	"context"
	"io"
	"net"
	"time"
)

// Reader is an interface that consists of a number of methods for reading that Conn must implement.
//
// Note that the methods in this interface are not concurrency-safe for concurrent use,
// you must invoke them within any method in EventHandler.
type Reader interface {
	io.Reader
	io.WriterTo

	// Next returns the next n bytes and advances the inbound buffer.
	// buf must not be used in a new goroutine. Otherwise, use Read instead.
	//
	// If the number of the available bytes is less than requested,
	// a pair of (0, io.ErrShortBuffer) is returned.
	Next(n int) (buf []byte, err error)

	// Peek returns the next n bytes without advancing the inbound buffer,
	// the returned bytes remain valid until a Discard is called.
	// buf must neither be used in a new goroutine nor anywhere after the call
	// to Discard, make a copy of buf manually or use Read otherwise.
	//
	// If the number of the available bytes is less than requested,
	// a pair of (0, io.ErrShortBuffer) is returned.
	Peek(n int) (buf []byte, err error)

	// Discard advances the inbound buffer with next n bytes, returning the number of bytes discarded.
	Discard(n int) (discarded int, err error)

	// InboundBuffered returns the number of bytes that can be read from the current buffer.
	InboundBuffered() int
}

// Writer is an interface that consists of a number of methods for writing that Conn must implement.
type Writer interface {
	io.Writer     // not concurrency-safe
	io.ReaderFrom // not concurrency-safe

	// SendTo transmits a message to the given address, it's not concurrency-safe.
	// It is available only for UDP sockets, an ErrUnsupportedOp will be returned
	// when it is called on a non-UDP socket.
	// This method should be used only when you need to send a message to a specific
	// address over the UDP socket, otherwise you should use Conn.Write() instead.
	SendTo(buf []byte, addr net.Addr) (n int, err error)

	// Writev writes multiple byte slices to remote synchronously, it's not concurrency-safe,
	// you must invoke it within any method in EventHandler.
	Writev(bs [][]byte) (n int, err error)

	// Flush writes any buffered data to the underlying connection, it's not concurrency-safe,
	// you must invoke it within any method in EventHandler.
	Flush() error

	// OutboundBuffered returns the number of bytes that can be read from the current buffer.
	// it's not concurrency-safe, you must invoke it within any method in EventHandler.
	OutboundBuffered() int

	// AsyncWrite writes bytes to remote asynchronously, it's concurrency-safe,
	// you don't have to invoke it within any method in EventHandler,
	// usually you would call it in an individual goroutine.
	//
	// Note that it will go synchronously with UDP, so it is needless to call
	// this asynchronous method, we may disable this method for UDP and just
	// return ErrUnsupportedOp in the future, therefore, please don't rely on
	// this method to do something important under UDP, if you're working with UDP,
	// just call Conn.Write to send back your data.
	AsyncWrite(buf []byte, callback AsyncCallback) (err error)

	// AsyncWritev writes multiple byte slices to remote asynchronously,
	// you don't have to invoke it within any method in EventHandler,
	// usually you would call it in an individual goroutine.
	AsyncWritev(bs [][]byte, callback AsyncCallback) (err error)
}

// AsyncCallback is a callback that will be invoked after the asynchronous function finishes.
//
// Note that the parameter gnet.Conn might have been already released when it's UDP protocol,
// thus it shouldn't be accessed.
// This callback will be executed in event-loop, thus it must not block, otherwise,
// it blocks the event-loop.
type AsyncCallback func(c Conn, err error) error

// Socket is a set of functions which manipulate the underlying file descriptor of a connection.
//
// Note that the methods in this interface are concurrency-safe for concurrent use,
// you don't have to invoke them within any method in EventHandler.
type Socket interface {
	// Gfd returns the gfd of socket.
	// Gfd() gfd.GFD

	// Fd returns the underlying file descriptor.
	Fd() int

	// Dup returns a copy of the underlying file descriptor.
	// It is the caller's responsibility to close fd when finished.
	// Closing c does not affect fd, and closing fd does not affect c.
	//
	// The returned file descriptor is different from the
	//  connection. Attempting to change the properties of the original
	// using this duplicate may or may not have the desired effect.
	Dup() (int, error)

	// SetReadBuffer sets the size of the operating system's
	// receive buffer associated with the connection.
	SetReadBuffer(size int) error

	// SetWriteBuffer sets the size of the operating system's
	// transmit buffer associated with the connection.
	SetWriteBuffer(size int) error

	// SetLinger sets the behavior of Close on a connection which still
	// has data waiting to be sent or to be acknowledged.
	//
	// If secs < 0 (the default), the operating system finishes sending the
	// data in the background.
	//
	// If secs == 0, the operating system discards any unsent or
	// unacknowledged data.
	//
	// If secs > 0, the data is sent in the background as with sec < 0. On
	// some operating systems after sec seconds have elapsed any remaining
	// unsent data may be discarded.
	SetLinger(secs int) error

	// SetKeepAlivePeriod tells the operating system to send keep-alive
	// messages on the connection and sets period between TCP keep-alive probes.
	SetKeepAlivePeriod(d time.Duration) error

	// SetKeepAlive enables/disables the TCP keepalive with all socket options:
	// TCP_KEEPIDLE, TCP_KEEPINTVL and TCP_KEEPCNT. idle is the value for TCP_KEEPIDLE,
	// intvl is the value for TCP_KEEPINTVL, cnt is the value for TCP_KEEPCNT,
	// ignored when enabled is false.
	//
	// With TCP keep-alive enabled, idle is the time (in seconds) the connection
	// needs to remain idle before TCP starts sending keep-alive probes,
	// intvl is the time (in seconds) between individual keep-alive probes.
	// TCP will drop the connection after sending cnt probes without getting
	// any replies from the peer; then the socket is destroyed, and OnClose
	// is triggered.
	//
	// If one of idle, intvl, or cnt is less than 1, an error is returned.
	SetKeepAlive(enabled bool, idle, intvl time.Duration, cnt int) error

	// SetNoDelay controls whether the operating system should delay
	// packet transmission in hopes of sending fewer packets (Nagle's
	// algorithm).
	// The default is true (no delay), meaning that data is sent as soon as possible after a Write.
	SetNoDelay(noDelay bool) error
}

// Conn is an interface of underlying connection.
type Conn interface {
	Reader // all methods in Reader are not concurrency-safe.
	Writer // some methods in Writer are concurrency-safe, some are not.
	Socket // all methods in Socket are concurrency-safe.

	// Context returns a user-defined context, it's not concurrency-safe,
	// you must invoke it within any method in EventHandler.
	Context() (ctx any)

	// EventLoop returns the event-loop that the connection belongs to.
	// The returned EventLoop is concurrency-safe.
	EventLoop() EventLoop

	// SetContext sets a user-defined context, it's not concurrency-safe,
	// you must invoke it within any method in EventHandler.
	SetContext(ctx any)

	// LocalAddr is the connection's local socket address, it's not concurrency-safe,
	// you must invoke it within any method in EventHandler.
	LocalAddr() net.Addr

	// RemoteAddr is the connection's remote address, it's not concurrency-safe,
	// you must invoke it within any method in EventHandler.
	RemoteAddr() net.Addr

	// Wake triggers an OnTraffic event for the current connection, it's concurrency-safe.
	Wake(callback AsyncCallback) error

	// CloseWithCallback closes the current connection, it's concurrency-safe.
	// Usually you should provide a non-nil callback for this method,
	// otherwise your better choice is Close().
	CloseWithCallback(callback AsyncCallback) error

	// Close closes the current connection, implements net.Conn, it's concurrency-safe.
	Close() error

	// SetDeadline implements net.Conn.
	SetDeadline(time.Time) error

	// SetReadDeadline implements net.Conn.
	SetReadDeadline(time.Time) error

	// SetWriteDeadline implements net.Conn.
	SetWriteDeadline(time.Time) error
}

// Runnable defines the common protocol of an execution on an event-loop.
// This interface should be implemented and passed to an event-loop in some way,
// then the event-loop will invoke Run to perform the execution.
// !!!Caution: Run must not contain any blocking operations like heavy disk or
// network I/O, or else it will block the event-loop.
type Runnable interface {
	// Run is about to be executed by the event-loop.
	Run(ctx context.Context) error
}

// RunnableFunc is an adapter to allow the use of ordinary function as a Runnable.
type RunnableFunc func(ctx context.Context) error

// Run executes the RunnableFunc itself.
func (fn RunnableFunc) Run(ctx context.Context) error {
	return fn(ctx)
}

// RegisteredResult is the result of a Register call.
type RegisteredResult struct {
	Conn Conn
	Err  error
}

// EventLoop provides a set of methods for manipulating the event-loop.
type EventLoop interface {
	// Register connects to the given address and registers the connection to the current event-loop,
	// it's concurrency-safe.
	Register(ctx context.Context, addr net.Addr) (<-chan RegisteredResult, error)
	// Enroll is like Register, but it accepts an established net.Conn instead of a net.Addr,
	// it's concurrency-safe.
	Enroll(ctx context.Context, c net.Conn) (<-chan RegisteredResult, error)
	// Execute will execute the given runnable on the event-loop at some time in the future,
	// it's concurrency-safe.
	Execute(ctx context.Context, runnable Runnable) error
	// Schedule is like Execute, but it allows you to specify when the runnable is executed.
	// In other words, the runnable will be executed when the delay duration is reached,
	// it's concurrency-safe.
	// TODO(panjf2000): not supported yet, implement this.
	Schedule(ctx context.Context, runnable Runnable, delay time.Duration) error

	// Close closes the given Conn that belongs to the current event-loop.
	// It must be called on the same event-loop that the connection belongs to.
	// This method is not concurrency-safe, you must invoke it on the event loop.
	Close(Conn) error
}
