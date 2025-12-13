package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go-antinuke-2.0/internal/sys"
)

type GatewayReader struct {
	conn          *websocket.Conn
	token         string
	intents       int
	sequenceNum   uint64
	sessionID     string
	heartbeatTick *time.Ticker
	eventQueue    *RingBuffer
	ctx           context.Context
	cancel        context.CancelFunc
	cpuCore       int
}

func NewGatewayReader(token string, intents int, eventQueue *RingBuffer, cpuCore int) *GatewayReader {
	ctx, cancel := context.WithCancel(context.Background())
	return &GatewayReader{
		token:      token,
		intents:    intents,
		eventQueue: eventQueue,
		ctx:        ctx,
		cancel:     cancel,
		cpuCore:    cpuCore,
	}
}

func (g *GatewayReader) Connect() error {
	runtime.LockOSThread()

	// Optimize websocket connection with larger buffers
	dialer := &websocket.Dialer{
		ReadBufferSize:  65536, // 64KB read buffer
		WriteBufferSize: 32768, // 32KB write buffer
		Proxy:           nil,
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial("wss://gateway.discord.gg/?v=10&encoding=json", nil)
	if err != nil {
		return err
	}
	g.conn = conn

	_, msg, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	var hello struct {
		Op int `json:"op"`
		D  struct {
			HeartbeatInterval int `json:"heartbeat_interval"`
		} `json:"d"`
	}

	if err := json.Unmarshal(msg, &hello); err != nil {
		return err
	}

	g.heartbeatTick = time.NewTicker(time.Duration(hello.D.HeartbeatInterval) * time.Millisecond)
	go g.heartbeatLoop()

	identify := map[string]interface{}{
		"op": 2,
		"d": map[string]interface{}{
			"token":   g.token,
			"intents": g.intents,
			"properties": map[string]string{
				"$os":      "linux",
				"$browser": "antinuke",
				"$device":  "antinuke",
			},
		},
	}

	return conn.WriteJSON(identify)
}

func (g *GatewayReader) heartbeatLoop() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case <-g.heartbeatTick.C:
			seq := atomic.LoadUint64(&g.sequenceNum)
			payload := map[string]interface{}{
				"op": 1,
				"d":  seq,
			}
			g.conn.WriteJSON(payload)
		}
	}
}

func (g *GatewayReader) ReadLoop() error {
	if err := sys.PinToCore(g.cpuCore); err != nil {
		fmt.Printf("Failed to pin ingest thread to core %d: %v\n", g.cpuCore, err)
	}
	runtime.LockOSThread()

	for {
		select {
		case <-g.ctx.Done():
			return nil
		default:
		}

		messageType, data, err := g.conn.ReadMessage()
		if err != nil {
			if err == io.EOF || websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return err
			}
			continue
		}

		if messageType != websocket.TextMessage {
			continue
		}

		g.processMessage(data)
	}
}

func (g *GatewayReader) processMessage(data []byte) {
	op := ExtractOp(data)

	// Update sequence number if present
	s := ExtractSeq(data)
	if s > 0 {
		atomic.StoreUint64(&g.sequenceNum, s)
	}

	// Dispatch event
	if op == 0 {
		t := ExtractType(data)
		if t != "" {
			d := ExtractData(data)
			if d != nil {
				event := SliceEvent(t, json.RawMessage(d))
				if event != nil && event.Priority >= 2 {
					g.eventQueue.Enqueue(event)
				}
			}
		}
	}
}

func (g *GatewayReader) Close() error {
	g.cancel()
	if g.heartbeatTick != nil {
		g.heartbeatTick.Stop()
	}
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

func (g *GatewayReader) Send(payload interface{}) error {
	return g.conn.WriteJSON(payload)
}

type HTTPGatewayInfo struct {
	URL    string `json:"url"`
	Shards int    `json:"shards"`
}

func GetGatewayInfo(token string) (*HTTPGatewayInfo, error) {
	req, _ := http.NewRequest("GET", "https://discord.com/api/v10/gateway/bot", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", token))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info HTTPGatewayInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}
