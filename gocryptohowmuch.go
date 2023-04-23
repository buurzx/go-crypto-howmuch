package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/buurzx/go-crypto-howmuch/ui"
	"github.com/buurzx/go-crypto-howmuch/websockets"
)

const (
	streamURL = "wss://stream.binance.com:9443/stream"
)

func main() {
	var (
		symbol     string
		symbolBase string
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// channels
	prices := make(chan float64)
	defer close(prices)

	done := make(chan struct{})
	defer close(done)

	errorCh := make(chan error, 1)
	defer close(errorCh)

	sigs := make(chan os.Signal, 1)
	defer close(sigs)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// flags
	flag.StringVar(&symbol, "symbol", "btc", "lower case crypto name")
	flag.StringVar(&symbolBase, "symbolbase", "usdt", "base, lower case crypto name")
	flag.Parse()

	// ws connection
	conn, err := websockets.NewWS(streamURL)
	if err != nil {
		panic(err)
	}

	appUI := ui.NewUI()
	symbolText := fmt.Sprintf(" %s / %s ", symbol, symbolBase)

	go appUI.StartRendering(ctx, symbolText, prices, errorCh, done)

	go conn.Listen(symbol, symbolBase, prices, errorCh)

	// blocks and wait signal
	select {
	case sig := <-sigs:
		log.Printf("Closing by signal ... %v \n", sig)
	case err := <-errorCh:
		log.Println("Something went wrong ", err.Error())
	case <-done:
		log.Println("Closing by done ch ... ")
	}

	// close ws conn
	if err = conn.Close(); err != nil {
		log.Println("ws connection failed to close")
	}

	time.Sleep(time.Second * 1)
}
