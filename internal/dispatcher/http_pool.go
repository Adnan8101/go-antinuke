package dispatcher

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type HTTPPool struct {
	clients []*http.Client
	size    int
	index   int
}

func NewHTTPPool(size int) *HTTPPool {
	clients := make([]*http.Client, size)

	dialer := &net.Dialer{
		Timeout:   2 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   50,
		MaxConnsPerHost:       50,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		},
		DisableCompression: true,
		DisableKeepAlives:  false,
		ForceAttemptHTTP2:  true,
	}

	for i := 0; i < size; i++ {
		clients[i] = &http.Client{
			Transport: transport,
			Timeout:   3 * time.Second,
		}
	}

	return &HTTPPool{
		clients: clients,
		size:    size,
		index:   0,
	}
}

func (hp *HTTPPool) GetClient() *http.Client {
	client := hp.clients[hp.index]
	hp.index = (hp.index + 1) % hp.size
	return client
}

func (hp *HTTPPool) Warmup() {
	warmupURL := "https://discord.com/api/v10/gateway"

	successCount := 0
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", warmupURL, nil)
		resp, err := hp.clients[0].Do(req)
		if err == nil && resp != nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				successCount++
			}
		}
		if successCount >= 3 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)
}
