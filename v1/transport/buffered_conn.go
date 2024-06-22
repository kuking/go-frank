package transport

import (
	"bufio"
	"net"
)

type BufferedConn struct {
	r *bufio.Reader
	w *bufio.Writer
	net.Conn
}

func NewBufferedConn(c net.Conn) BufferedConn {
	return BufferedConn{bufio.NewReader(c), bufio.NewWriter(c), c}
}

func NewBufferedConnSize(c net.Conn, n int) BufferedConn {
	return BufferedConn{bufio.NewReaderSize(c, n), bufio.NewWriterSize(c, n), c}
}

func (b BufferedConn) Peek(n int) ([]byte, error) {
	return b.r.Peek(n)
}

func (b BufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

func (b BufferedConn) Write(p []byte) (int, error) {
	return b.w.Write(p)
}

func (b *BufferedConn) WriteByte(c byte) error {
	return b.w.WriteByte(c)
}

func (b *BufferedConn) Flush() error {
	return b.w.Flush()
}
func (b *BufferedConn) WriteRune(r rune) (size int, err error) {
	return b.w.WriteRune(r)
}

func (b *BufferedConn) WriteString(s string) (int, error) {
	return b.w.WriteString(s)
}
