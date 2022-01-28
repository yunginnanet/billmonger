package invoice

import (
	"bytes"
	"fmt"
	"math"
	"path"
	"regexp"
	"strconv"
	"text/template"

	"github.com/jinzhu/now"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"
)

var InvoiceNumber string

type BusinessDetails struct {
	Name      string `yaml:"name"`
	Person    string `yaml:"person"`
	Address   string `yaml:"address"`
	ImageFile string `yaml:"image_file"`
	SansFont  string `yaml:"sans_font"`
	SerifFont string `yaml:"serif_font"`
}

type BillDetails struct {
	Department   string `yaml:"department"`
	Currency     string `yaml:"currency"`
	PaymentTerms string `yaml:"payment_terms"`
	DueDate      string `yaml:"due_date"`
	Date         string `yaml:"date"`
	UseExactDate bool   `yaml:"use_exact_date"`
}

func (b *BillDetails) Strings() []string {
	return []string{
		b.Department, b.Currency, b.PaymentTerms, b.DueDate,
	}
}

type BillToDetails struct {
	Email        string
	Name         string
	Street       string
	CityStateZip string `yaml:"city_state_zip"`
	Country      string
}

type BillableItem struct {
	Quantity    float64
	Description string
	UnitPrice   float64 `yaml:"unit_price"`
	Currency    string
}

func (b *BillableItem) Total() float64 {
	return b.UnitPrice * b.Quantity
}

func (b *BillableItem) Strings() []string {
	return []string{
		strconv.FormatFloat(b.Quantity, 'f', 2, 64),
		b.Description,
		b.Currency + " " + niceFloatStr(b.UnitPrice),
		b.Currency + " " + niceFloatStr(b.Total()),
	}
}

type TaxDetails struct {
	DefaultPercentage float64 `yaml:"default_percentage"`
	TaxName           string  `yaml:"tax_name"`
}

type BankDetails struct {
	TransferType string `yaml:"transfer_type"`
	Name         string
	// Address      string
	AccountType   string `yaml:"account_type"`
	RoutingNumber string `yaml:"routing_number"`
	AccountNumber string `yaml:"account_number"`
	// IBAN         string
	// SortCode     string `yaml:"sort_code"`
	// SWIFTBIC     string `yaml:"swift_bic"`
}

func (b *BankDetails) Strings() []string {
	return []string{
		b.TransferType, b.Name, b.AccountType, b.AccountNumber, b.RoutingNumber,
	}
}

type Color struct {
	R int
	G int
	B int
}

type BillColor struct {
	ColorLight Color `yaml:"color_light"`
	ColorDark  Color `yaml:"color_dark"`
}

type AppConfig struct {
	OutputDir string `yaml:"output_dir"`
}

type BillingConfig struct {
	Business  *BusinessDetails `yaml:"business"`
	Bill      *BillDetails     `yaml:"bill"`
	BillTo    *BillToDetails   `yaml:"bill_to"`
	Billables []BillableItem   `yaml:"billables"`
	Tax       *TaxDetails      `yaml:"tax"`
	Bank      *BankDetails     `yaml:"bank"`
	Colors    *BillColor       `yaml:"colors"`
	App       *AppConfig       `yaml:"app_config"`
}

// ParseConfig parses the YAML config file which contains the settings for the
// bill we're going to process. It uses a simple FuncMap to template the text,
// allowing the billing items to describe the current date range.
func ParseConfig(filename, billingDate, outputDir, invoiceNumber string) (*BillingConfig, error) {
	billTime := now.New(now.MustParse(billingDate))

	InvoiceNumber = invoiceNumber

	funcMap := template.FuncMap{
		"endOfNextMonth": func() string {
			return billTime.EndOfMonth().AddDate(0, 1, -1).Format("01/02/06")
		},
		"endOfThisMonth": func() string {
			return billTime.EndOfMonth().Format("01/02/06")
		},
		"billingPeriod": func() string {
			return billTime.BeginningOfMonth().Format("Jan 2, 2006") +
				" - " + billTime.EndOfMonth().Format("Jan 2, 2006")
		},
	}

	t, err := template.New("billing.yaml").Funcs(funcMap).ParseFiles(filename)
	if err != nil {
		return nil, fmt.Errorf("Error Parsing template '%s': %s", filename, err.Error())
	}

	buf := bytes.NewBuffer(make([]byte, 0, 65535))
	err = t.ExecuteTemplate(buf, path.Base(filename), nil)
	if err != nil {
		return nil, err
	}

	var config BillingConfig
	err = yaml.Unmarshal(buf.Bytes(), &config)
	if err != nil {
		return nil, err
	}

	// Set the date we'll bill on
	config.Bill.Date = billingDate

	// Set the output dir if it's not configured in the file
	if config.App == nil || (config.App != nil && config.App.OutputDir == "") {
		config.App = &AppConfig{OutputDir: outputDir}
	}

	return &config, nil
}

// niceFloatStr takes a float and gives back a monetary, human-formatted
// value.
func niceFloatStr(f float64) string {
	roundedFloat := math.Round(f*100) / 100
	r := regexp.MustCompile("-?[0-9,]+.[0-9]{2}")
	p := message.NewPrinter(language.English)
	results := r.FindAllString(p.Sprintf("%f", roundedFloat), 1)

	if len(results) < 1 {
		panic("got some ridiculous number that has no decimals")
	}

	return results[0]
}
