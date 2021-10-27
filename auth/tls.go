package auth

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc/credentials"
)

// GetServerTlsCredentials creates an appropriate TLS configuration for server-side use.
func GetServerTlsCredentials(certPath string, certKeyPath string, caCertPath string) (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(certPath, certKeyPath)
	if err != nil {
		return nil, err
	}

	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	config := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		Certificates: []tls.Certificate{serverCert},
		MinVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}

	return credentials.NewTLS(config), nil
}

// GetClientTlsCredentials creates an appropriate TLS configuration for client-side use.
func GetClientTlsCredentials(certPath string, certKeyPath string, caCertPath string) (credentials.TransportCredentials, error) {
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	clientCert, err := tls.LoadX509KeyPair(certPath, certKeyPath)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}

	return credentials.NewTLS(config), nil
}
