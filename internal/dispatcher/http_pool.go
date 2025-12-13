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
			// Absolute minimum latency settings
			MaxConnsPerHost:     500,     // Increased for more parallel connections
			MaxIdleConnDuration: 90 * time.Second,
			MaxConnDuration:     0,       // Unlimited connection duration
			ReadTimeout:         1 * time.Second,  // Faster timeout
			WriteTimeout:        1 * time.Second,  // Faster timeout
			MaxConnWaitTimeout:  500 * time.Millisecond,  // Reduced wait time

			// Maximum performance optimizations
			ReadBufferSize:      16384,   // Increased buffer
			WriteBufferSize:     16384,   // Increased buffer
			MaxResponseBodySize: 2 * 1024 * 1024, // 2MB

			// Speed optimizations
			DisableHeaderNamesNormalizing: true,  // Skip header normalization for speed
			DisablePathNormalizing:        true,  // Skip path normalization for speed
			
			// No retries for minimum latency
			MaxIdemponentCallAttempts: 1,
			
			// Dial settings for fastest connection
			DialDualStack: true,  // Try IPv4 and IPv6 simultaneously
			
			// TLS config - use default for now
			TLSConfig: nil,
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
