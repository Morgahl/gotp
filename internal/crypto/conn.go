package crypto

import (
	"crypto/tls"
	"errors"
	"net"
)

func Listen(addr string, cert tls.Certificate) (net.Listener, error) {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		var netErr *net.AddrError
		// TODO: string matching an error here?
		if errors.As(err, &netErr) && netErr.Err == "missing port in address" {
			// TODO: We will need to define a range via config/env later on
			addr = net.JoinHostPort(addr, "0")
		} else {
			return nil, err
		}
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// we disable verification for simplicity, but this is insecure. We do a simple HMAC handshake with a nonce and
		// the cookie is at least the expected local cookie. It is generally up to the environmental config to ensure
		// that deployments with the same cookie are not in the same network if they are not supposed to be.
		InsecureSkipVerify: true,
	}

	return tls.Listen("tcp", addr, config)
}

func Dial(addr string, cert tls.Certificate) (*tls.Conn, error) {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// we disable verification for simplicity, but this is insecure. We do a simple HMAC handshake with a nonce and
		// the cookie is at least the expected local cookie. It is generally up to the environmental config to ensure
		// that deployments with the same cookie are not in the same network if they are not supposed to be.
		InsecureSkipVerify: true,
	}

	return tls.Dial("tcp", addr, tlsConfig)
}
