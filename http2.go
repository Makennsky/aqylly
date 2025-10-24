package aqylly

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// HTTP2Config holds HTTP/2 server configuration
type HTTP2Config struct {
	// MaxConcurrentStreams limits the number of concurrent streams per connection
	MaxConcurrentStreams uint32

	// MaxReadFrameSize is the maximum frame size for DATA frames
	MaxReadFrameSize uint32

	// IdleTimeout specifies how long idle connections should be kept alive
	IdleTimeout uint32

	// MaxUploadBufferPerConnection is the size of the flow control window for uploads
	MaxUploadBufferPerConnection int32

	// MaxUploadBufferPerStream is the size of the flow control window for each stream
	MaxUploadBufferPerStream int32
}

// DefaultHTTP2Config returns default HTTP/2 configuration
func DefaultHTTP2Config() *HTTP2Config {
	return &HTTP2Config{
		MaxConcurrentStreams:         250,
		MaxReadFrameSize:             16384, // 16 KB
		IdleTimeout:                  120,   // 2 minutes
		MaxUploadBufferPerConnection: 1 << 20, // 1 MB
		MaxUploadBufferPerStream:     1 << 20, // 1 MB
	}
}

// toHTTP2Server converts HTTP2Config to http2.Server
func (cfg *HTTP2Config) toHTTP2Server() *http2.Server {
	return &http2.Server{
		MaxConcurrentStreams:         cfg.MaxConcurrentStreams,
		MaxReadFrameSize:             cfg.MaxReadFrameSize,
		IdleTimeout:                  0, // handled by http.Server
		MaxUploadBufferPerConnection: cfg.MaxUploadBufferPerConnection,
		MaxUploadBufferPerStream:     cfg.MaxUploadBufferPerStream,
	}
}

// ConfigureHTTP2Server configures HTTP/2 for the server
func ConfigureHTTP2Server(srv *http.Server, cfg *HTTP2Config) error {
	if cfg == nil {
		cfg = DefaultHTTP2Config()
	}

	h2s := cfg.toHTTP2Server()
	return http2.ConfigureServer(srv, h2s)
}

// NewH2CHandler creates an HTTP/2 Cleartext handler
// This is useful for internal microservices that don't need TLS
func NewH2CHandler(handler http.Handler, cfg *HTTP2Config) http.Handler {
	if cfg == nil {
		cfg = DefaultHTTP2Config()
	}

	h2s := cfg.toHTTP2Server()
	return h2c.NewHandler(handler, h2s)
}

// ConfigureTLSForHTTP2 configures TLS for HTTP/2 with ALPN
func ConfigureTLSForHTTP2() *tls.Config {
	return &tls.Config{
		NextProtos: []string{"h2", "http/1.1"},
		MinVersion: tls.VersionTLS12,
	}
}

// NewHTTP2Transport creates a new HTTP/2 transport for clients
func NewHTTP2Transport() *http.Transport {
	return &http.Transport{
		ForceAttemptHTTP2: true,
		TLSClientConfig:   ConfigureTLSForHTTP2(),
	}
}

// NewH2CTransport creates an HTTP/2 Cleartext transport for clients
func NewH2CTransport() *http.Transport {
	return &http.Transport{
		ForceAttemptHTTP2: true,
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Use context-aware dial for h2c
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}
}
