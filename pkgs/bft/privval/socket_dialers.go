package privval

import (
	"net"
	"time"

	"github.com/gnolang/gno/pkgs/crypto/ed25519"
	"github.com/gnolang/gno/pkgs/errors"
	osm "github.com/gnolang/gno/pkgs/os"
	p2pconn "github.com/gnolang/gno/pkgs/p2p/conn"
)

// Socket errors.
var (
	ErrDialRetryMax = errors.New("dialed maximum retries")
)

// SocketDialer dials a remote address and returns a net.Conn or an error.
type SocketDialer func() (net.Conn, error)

// DialTCPFn dials the given tcp addr, using the given timeoutReadWrite and
// privKey for the authenticated encryption handshake.
func DialTCPFn(addr string, timeoutReadWrite time.Duration, privKey ed25519.PrivKeyEd25519) SocketDialer {
	return func() (net.Conn, error) {
		conn, err := osm.Connect(addr)
		if err == nil {
			deadline := time.Now().Add(timeoutReadWrite)
			err = conn.SetDeadline(deadline)
		}
		if err == nil {
			conn, err = p2pconn.MakeSecretConnection(conn, privKey)
		}
		return conn, err
	}
}

// DialUnixFn dials the given unix socket.
func DialUnixFn(addr string) SocketDialer {
	return func() (net.Conn, error) {
		unixAddr := &net.UnixAddr{Name: addr, Net: "unix"}
		return net.DialUnix("unix", nil, unixAddr)
	}
}
