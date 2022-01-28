package main

import (
	"os"
	"time"

	"github.com/jinzhu/now"
	"github.com/relistan/billmonger/invoice"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/relistan/rubberneck.v1"
)

const (
	sansFont  = "Helvetica"
	serifFont = "Times"
)

type CliConfig struct {
	ConfigFile    *string
	BillingDate   *string
	OutputDir     *string
	InvoiceNumber *string
}

func checkImageFile(config *invoice.BillingConfig) error {
	_, err := os.Stat(config.Business.ImageFile)
	return err
}

func main() {
	cli := CliConfig{
		ConfigFile: kingpin.Flag("config-file", "The YAML config file to use").Short('c').Default("billing.yaml").String(),
		BillingDate: kingpin.Flag("billing-date", "The date to assume the bill is written on").
			Short('b').Default(time.Now().Format("2006-01-02")).String(),
		OutputDir:     kingpin.Flag("output-dir", "The output directory to use. Overridden by config file.").Short('o').Default(".").String(),
		InvoiceNumber: kingpin.Flag("invoice-number", "The invoice number.").Short('i').Default(time.Now().Format("Jan22006")).String(),
	}
	kingpin.Parse()

	// Make sure the supplied time is a valid one
	_, err := now.Parse(*cli.BillingDate)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	config, err := invoice.ParseConfig(*cli.ConfigFile, *cli.BillingDate, *cli.OutputDir, *cli.InvoiceNumber)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	// Print the config
	printer := rubberneck.NewDefaultPrinter()
	printer.PrintWithLabel("Settings ("+*cli.ConfigFile+")", config)

	// Pick up some defaults where needed
	if config.Business.SansFont == "" {
		config.Business.SansFont = sansFont
	}

	if config.Business.SerifFont == "" {
		config.Business.SerifFont = serifFont
	}

	err = checkImageFile(config)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	bill := invoice.NewBill(config)

	err = bill.RenderToFile()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
