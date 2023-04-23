package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/rivo/tview"
	"golang.org/x/term"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	redColor   = "red"
	greenColor = "green"
)

var tickerDuration = 300 * time.Millisecond

type UI struct {
	app      *tview.Application
	textView *tview.TextView
}

func NewUI() *UI {
	ui := &UI{app: tview.NewApplication()}

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText("").
		SetWrap(false).
		SetChangedFunc(func() {
			ui.app.Draw()
		})

	ui.textView = textView

	return ui
}

func (ui *UI) StartRendering(
	ctx context.Context,
	symbolText string,
	prices chan float64,
	errorCH chan error,
	done chan struct{},
) {
	plot := ""
	price := float64(0)
	prevPrice := float64(0)
	graphCutSize := 10
	width, _, _ := term.GetSize(0)

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	printer := message.NewPrinter(language.English)

	graphWidth := width - graphCutSize
	dataGraph := make([]float64, 0, graphWidth)

	go func() {
		ticker := time.NewTicker(tickerDuration)
		defer ticker.Stop()

	loop:
		for {
			select {
			case <-childCtx.Done():
				break loop
			case price = <-prices:
				select {
				case <-ticker.C:
					dataGraph = ui.appendGraph(graphWidth, dataGraph, price)
					plot = asciigraph.Plot(dataGraph, asciigraph.Precision(1), asciigraph.Height(10))

					ui.app.QueueUpdateDraw(
						func() {
							colorText := redColor

							if price > prevPrice {
								colorText = greenColor
							}

							ui.textView.SetText(
								printer.Sprintf("   %s: [%s]%.3f[default]\n\n%v", symbolText, colorText, price, plot),
							)
						},
					)

					prevPrice = price
				default:
				}
			}
		}
	}()

	if err := ui.app.SetRoot(ui.textView, true).EnableMouse(true).Run(); err != nil {
		errorCH <- fmt.Errorf("failed set root %w", err)
	}

	done <- struct{}{}
}

func (ui *UI) appendGraph(size int, data []float64, newItem float64) []float64 {
	if len(data) < size {
		return append(data, newItem)
	}

	copy(data[0:], data[len(data)-size+1:])
	data[size-1] = newItem
	return data[:size-1]
}
