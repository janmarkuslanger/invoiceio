package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"

	"github.com/janmarkuslanger/invoiceio/internal/storage"
	"github.com/janmarkuslanger/invoiceio/internal/ui"
)

func main() {
	dataDirFlag := flag.String("data", "data", "directory used to store application data")
	flag.Parse()

	dataDir, err := filepath.Abs(*dataDirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve data directory: %v\n", err)
		os.Exit(1)
	}

	store, err := storage.New(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialise storage: %v\n", err)
		os.Exit(1)
	}

	app := fyneApp.NewWithID("invoiceio")
	window := app.NewWindow("InvoiceIO")
	uiLayer := ui.New(store, window)
	window.SetContent(uiLayer.Build())
	window.Resize(fyne.NewSize(960, 640))
	window.ShowAndRun()
}
