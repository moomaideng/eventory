package receiptpdf

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/moomaideng/eventory/services/payment/models"
	"github.com/signintech/gopdf"
)

// shortcut: font covers Thai and Latin; add a fallback if other scripts are needed.
//
//go:embed fonts/Sarabun-Regular.ttf
var font []byte

//go:embed fonts/Sarabun-Bold.ttf
var boldFont []byte

func Generate(invoice *models.Invoice) ([]byte, error) {
	if invoice.PaidAt == nil || invoice.Receipt == nil {
		return nil, fmt.Errorf("receipt requires a paid invoice")
	}
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()
	if err := pdf.AddTTFFontData("Sarabun", font); err != nil {
		return nil, err
	}
	if err := pdf.AddTTFFontData("SarabunBold", boldFont); err != nil {
		return nil, err
	}
	pdf.SetInfo(gopdf.PdfInfo{Title: "Eventory receipt", Author: "Eventory", Subject: "Course-project demonstration receipt"})
	var drawErr error
	// Keep wrapping and PDF errors in one place; positions are A4 points.
	write := func(text string, x, y, width float64, size float64, bold bool) float64 {
		if drawErr != nil {
			return y
		}
		family := "Sarabun"
		if bold {
			family = "SarabunBold"
		}
		if drawErr = pdf.SetFont(family, "", size); drawErr != nil {
			return y
		}
		var lines []string
		lines, drawErr = pdf.SplitText(strings.Join(strings.Fields(text), " "), width)
		if drawErr != nil {
			return y
		}
		for _, line := range lines {
			pdf.SetXY(x, y)
			if drawErr = pdf.Text(line); drawErr != nil {
				return y
			}
			y += size + 5
		}
		return y
	}

	pdf.SetFillColor(49, 112, 62)
	if err := pdf.Rectangle(44, 40, 551, 122, "F", 10, 8); err != nil {
		return nil, err
	}
	pdf.SetTextColor(255, 255, 255)
	write("Eventory.", 64, 78, 240, 26, true)
	write("PAYMENT RECEIPT", 365, 72, 170, 11, true)
	write("Tournament management", 64, 102, 240, 10, false)

	pdf.SetTextColor(49, 112, 62)
	write("Payment confirmation", 44, 157, 507, 13, true)
	pdf.SetTextColor(104, 116, 107)
	write("Keep this receipt for your records.", 44, 178, 507, 10, false)

	pdf.SetFillColor(240, 246, 240)
	if err := pdf.Rectangle(44, 203, 551, 298, "F", 10, 8); err != nil {
		return nil, err
	}
	pdf.SetTextColor(49, 112, 62)
	write("TOTAL PAID", 64, 232, 300, 10, true)
	write("PAID", 434, 232, 100, 10, true)
	pdf.SetTextColor(26, 45, 31)
	write(formatAmount(invoice), 64, 277, 467, 30, true)

	write("Receipt details", 44, 329, 507, 13, true)
	pdf.SetStrokeColor(223, 230, 223)
	pdf.SetLineWidth(0.7)
	pdf.Line(44, 338, 551, 338)
	for _, field := range []struct {
		label, value string
		x, y, width  float64
	}{
		{"Receipt ID", invoice.Receipt.ID.String(), 44, 369, 240},
		{"Paid at (UTC)", invoice.PaidAt.UTC().Format("02 Jan 2006, 15:04:05 UTC"), 318, 369, 233},
		{"Invoice ID", invoice.ID.String(), 44, 421, 240},
		{"Currency", invoice.Currency, 318, 421, 233},
		{"Payer account", invoice.PayerAccountID.String(), 44, 473, 507},
	} {
		pdf.SetTextColor(104, 116, 107)
		write(field.label, field.x, field.y-18, field.width, 9, false)
		pdf.SetTextColor(26, 45, 31)
		write(field.value, field.x, field.y, field.width, 10.5, false)
	}
	write("Payment details", 44, 511, 507, 13, true)
	pdf.Line(44, 525, 551, 525)
	y := 548.0
	for _, field := range []struct{ label, value string }{
		{"Source", invoice.Source}, {"Reference", invoice.Reference}, {"Description", invoice.Description},
	} {
		pdf.SetTextColor(104, 116, 107)
		write(field.label, 44, y, 507, 9, false)
		pdf.SetTextColor(26, 45, 31)
		y = write(field.value, 44, y+15, 507, 10.5, false) + 7
	}
	if y > 790 {
		return nil, fmt.Errorf("receipt details exceed the page")
	}
	pdf.Line(44, 790, 551, 790)
	pdf.SetTextColor(104, 116, 107)
	write("Course-project demo · No money transferred · Not proof of payment", 44, 811, 507, 9, false)
	if drawErr != nil {
		return nil, drawErr
	}
	return pdf.GetBytesPdfReturnErr()
}

func formatAmount(invoice *models.Invoice) string {
	whole := strconv.FormatInt(invoice.AmountMinor/100, 10)
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}
	return fmt.Sprintf("%s %s.%02d", invoice.Currency, whole, invoice.AmountMinor%100)
}
