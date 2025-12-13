package dispatcher

import (
	"crypto/tls"
	"time"

	"github.com/valyala/fasthttp"
)

type HTTPPool struct {
	clients []*fasthttp.Client
	size    int
	index   int
}

func NewHTTPPool(size int) *HTTPPool {
	clients := make([]*fasthttp.Client, size)

	// Ultra-aggressive TLS config for minimum latency
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
		// Aggressive cipher suites for speed
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		// Session cache for faster reconnections
		ClientSessionCache: tls.NewLRUClientSessionCache(128),
		// Reuse sessions
		SessionTicketsDisabled: false,
	}

	for i := 0; i < size; i++ {
		clients[i] = &fasthttp.Client{
			// EXTREME performance settings
			MaxConnsPerHost:     2000, // Maximum connections
			MaxIdleConnDuration: 180 * time.Second,
			MaxConnDuration:     0,                      // Never close connections
			ReadTimeout:         600 * time.Millisecond, // Ultra-fast timeout
			WriteTimeout:        600 * time.Millisecond,
			MaxConnWaitTimeout:  150 * time.Millisecond, // Minimum wait

			// Maximum buffer sizes for throughput
			ReadBufferSize:      65536,           // 64KB
			WriteBufferSize:     65536,           // 64KB
			MaxResponseBodySize: 4 * 1024 * 1024, // 4MB

			// Skip ALL normalization for raw speed
			DisableHeaderNamesNormalizing: true,
			DisablePathNormalizing:        true,

			// No retries - fail fast
			MaxIdemponentCallAttempts: 1,

			// Connection optimizations
			DialDualStack: true, // IPv4 + IPv6 simultaneously

			// TLS optimization
			TLSConfig: tlsConfig,

			// Disable compression for speed
			NoDefaultUserAgentHeader: true,
		}
	}

	return &HTTPPool{
		clients: clients,
		size:    size,
		index:   0,
	}
}

func (hp *HTTPPool) GetClient() *fasthttp.Client {
	client := hp.clients[hp.index]
	hp.index = (hp.index + 1) % hp.size
	return client
}

func (hp *HTTPPool) Warmup() {
	warmupURL := "https://discord.com/api/v10/gateway"

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	successCount := 0
	for i := 0; i < 3; i++ {
		req.SetRequestURI(warmupURL)
		req.Header.SetMethod("GET")

		err := hp.clients[0].DoTimeout(req, resp, 2*time.Second)
		if err == nil && resp.StatusCode() == 200 {
			successCount++
		}

		if successCount >= 2 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)
}
