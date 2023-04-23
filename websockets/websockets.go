package websockets

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
)

type WS struct {
	Conn *websocket.Conn
}

type AggMessage struct {
	Stream string `json:"stream"`
	Data   struct {
		P string `json:"p"`
	} `json:"data"`
}

func NewWS(streamURL string) (WS, error) {
	url, err := url.Parse(streamURL)
	if err != nil {
		return WS{}, fmt.Errorf("failed parse %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return WS{}, fmt.Errorf("failed dial ws conenction %w", err)
	}

	return WS{conn}, nil
}

func (ws WS) Close() error {
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	err := ws.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("failed to close ws connection %w", err)
	}

	return nil
}

func (ws WS) Listen(symbol, symbolBase string, prices chan float64, errorCh chan error) {
	err := ws.Conn.WriteMessage(websocket.TextMessage, []byte(`{"method": "SUBSCRIBE","params": ["`+symbol+symbolBase+`@aggTrade"],"id": 1}`))
	if err != nil {
		errorCh <- fmt.Errorf("failed write to conn %w", err)
		return
	}

loop:
	for {
		_, message, err := ws.Conn.ReadMessage()
		if err != nil {
			errorCh <- fmt.Errorf("failed read from conn %w", err)
			break loop
		}

		var msg AggMessage
		if err = json.Unmarshal(message, &msg); err != nil {
			errorCh <- fmt.Errorf("failed decode json %w", err)
		}

		price, err := strconv.ParseFloat(msg.Data.P, 64)
		if err != nil {
			// TODO print warning!
			continue

		}

		prices <- price
	}
}
