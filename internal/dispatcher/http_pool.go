package dispatcher

import (
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

	for i := 0; i < size; i++ {
		clients[i] = &fasthttp.Client{
			// Ultra-low latency settings
			MaxConnsPerHost:     200,
			MaxIdleConnDuration: 60 * time.Second,
			MaxConnDuration:     10 * time.Minute,
			ReadTimeout:         2 * time.Second,
			WriteTimeout:        2 * time.Second,
			MaxConnWaitTimeout:  1 * time.Second,

			// Performance optimizations
			ReadBufferSize:      8192,
			WriteBufferSize:     8192,
			MaxResponseBodySize: 1024 * 1024, // 1MB

			// Disable compression for speed
			DisableHeaderNamesNormalizing: false,
			DisablePathNormalizing:        false,

			// Keep connections alive
			MaxIdemponentCallAttempts: 1,

			// TLS config
			TLSConfig: nil, // Use default
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
