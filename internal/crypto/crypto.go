package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"time"

	"github.com/Morgahl/gotp"
)

const (
	NONCE_LENGTH = 64
	MAC_LENGTH   = 64
)

func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NONCE_LENGTH)
	_, err := rand.Read(nonce)
	return nonce, err
}

func ReadNonce(conn net.Conn) ([NONCE_LENGTH]byte, error) {
	var nonce [NONCE_LENGTH]byte
	n, err := conn.Read(nonce[:])
	if err != nil {
		return nonce, err
	}
	if n != NONCE_LENGTH {
		return nonce, io.ErrUnexpectedEOF
	}
	return nonce, nil
}

func ComputeHMAC(psk, nonce []byte) []byte {
	hmac := hmac.New(sha512.New, psk)
	hmac.Write(nonce)
	return hmac.Sum(nil)
}

func ReadHMAC(conn net.Conn) ([MAC_LENGTH]byte, error) {
	var mac [MAC_LENGTH]byte
	n, err := io.ReadFull(conn, mac[:])
	if err != nil {
		return mac, err
	}
	if n != MAC_LENGTH {
		return mac, io.ErrUnexpectedEOF
	}
	return mac, nil
}

func VerifyMAC(psk, nonce, receivedMAC []byte) bool {
	expectedMAC := ComputeHMAC(psk, nonce)
	return hmac.Equal(expectedMAC, receivedMAC)
}

func GenerateSelfSignedCert(host gotp.Atom, cookie gotp.Atom) (tls.Certificate, error) {
	hash := sha512.Sum512([]byte(host.String() + cookie.String()))
	serial := new(big.Int).SetBytes(hash[:20]) // use first 20 bytes (RFC max)

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: host.String(),
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // ~10 years
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return tls.X509KeyPair(certPEM, keyPEM)
}
