package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/janmarkuslanger/invoiceio/internal/i18n"
	"github.com/janmarkuslanger/invoiceio/internal/models"
)

const (
	pageWidth      = 595.28 // A4 in points
	pageHeight     = 841.89
	leftMargin     = 56.0
	topMargin      = 780.0
	lineHeight     = 16.0
	defaultFont    = "Courier"
	defaultFontRef = "/F1"
)

// CreateInvoicePDF renders a very small single page PDF invoice document.
func CreateInvoicePDF(outputPath string, profile models.Profile, customer models.Customer, invoice models.Invoice) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("pdf: ensure directory: %w", err)
	}

	lines := buildInvoiceLines(profile, customer, invoice)
	contentStream := buildContentStream(lines)
	doc, err := assembleSinglePagePDF(contentStream)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, doc, 0o644); err != nil {
		return fmt.Errorf("pdf: write file: %w", err)
	}
	return nil
}

func buildInvoiceLines(profile models.Profile, customer models.Customer, invoice models.Invoice) []string {
	now := time.Now().Format("2006-01-02 15:04")

	lines := []string{
		fmt.Sprintf("%s", strings.TrimSpace(profile.DisplayName)),
		strings.TrimSpace(profile.CompanyName),
		strings.TrimSpace(profile.AddressLine1),
		strings.TrimSpace(profile.AddressLine2),
		fmt.Sprintf("%s %s", strings.TrimSpace(profile.PostalCode), strings.TrimSpace(profile.City)),
		strings.TrimSpace(profile.Country),
		i18n.T("pdf.label.email", strings.TrimSpace(profile.Email)),
		i18n.T("pdf.label.phone", strings.TrimSpace(profile.Phone)),
		i18n.T("pdf.label.taxID", strings.TrimSpace(profile.TaxID)),
		"",
		i18n.T("pdf.label.invoiceNumber", invoice.Number),
		i18n.T("pdf.label.issuedOn", invoice.IssueDate.Format("2006-01-02")),
		i18n.T("pdf.label.dueDate", invoice.DueDate.Format("2006-01-02")),
		i18n.T("pdf.label.generatedOn", now),
		"",
		i18n.T("pdf.section.billTo"),
		strings.TrimSpace(customer.DisplayName),
		strings.TrimSpace(customer.ContactName),
		strings.TrimSpace(customer.AddressLine1),
		strings.TrimSpace(customer.AddressLine2),
		fmt.Sprintf("%s %s", strings.TrimSpace(customer.PostalCode), strings.TrimSpace(customer.City)),
		strings.TrimSpace(customer.Country),
		i18n.T("pdf.label.email", strings.TrimSpace(customer.Email)),
		i18n.T("pdf.label.phone", strings.TrimSpace(customer.Phone)),
		"",
		i18n.T("pdf.section.items"),
		fmt.Sprintf("%-40s %6s %10s %12s",
			i18n.T("pdf.items.column.description"),
			i18n.T("pdf.items.column.quantity"),
			i18n.T("pdf.items.column.unit"),
			i18n.T("pdf.items.column.lineTotal"),
		),
		strings.Repeat("-", 70),
	}

	for _, item := range invoice.Items {
		lines = append(lines, fmt.Sprintf("%-40s %6.2f %10.2f %12.2f", item.Description, item.Quantity, item.UnitPrice, item.LineTotal))
	}

	lines = append(lines,
		strings.Repeat("-", 70),
		fmt.Sprintf("%-40s %28.2f", i18n.T("pdf.label.subtotal"), invoice.Subtotal),
		fmt.Sprintf("%-40s %27.2f (%0.2f%%)", i18n.T("pdf.label.tax"), invoice.TaxAmount, invoice.TaxRatePercent),
		fmt.Sprintf("%-40s %28.2f", i18n.T("pdf.label.total"), invoice.Total),
	)

	if strings.TrimSpace(invoice.Notes) != "" {
		lines = append(lines, "", i18n.T("pdf.section.notes"))
		for _, line := range strings.Split(invoice.Notes, "\n") {
			lines = append(lines, line)
		}
	}

	lines = append(lines, "", i18n.T("pdf.section.paymentDetails"))
	if profile.PaymentDetails.BankName != "" {
		lines = append(lines, i18n.T("pdf.label.bank", profile.PaymentDetails.BankName))
	}
	if profile.PaymentDetails.IBAN != "" {
		lines = append(lines, i18n.T("pdf.label.iban", profile.PaymentDetails.IBAN))
	}
	if profile.PaymentDetails.BIC != "" {
		lines = append(lines, i18n.T("pdf.label.bic", profile.PaymentDetails.BIC))
	}
	if profile.PaymentDetails.PaymentTerms != "" {
		lines = append(lines, i18n.T("pdf.label.terms", profile.PaymentDetails.PaymentTerms))
	}

	return sanitizeLines(lines)
}

func sanitizeLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			out[i] = " "
			continue
		}
		out[i] = line
	}
	return out
}

func buildContentStream(lines []string) []byte {
	var buf bytes.Buffer
	buf.WriteString("BT\n")
	buf.WriteString(fmt.Sprintf("%s %0.2f Tf\n", defaultFontRef, 12.0))
	buf.WriteString(fmt.Sprintf("%0.2f TL\n", lineHeight))
	buf.WriteString(fmt.Sprintf("1 0 0 1 %0.2f %0.2f Tm\n", leftMargin, topMargin))
	for _, line := range lines {
		buf.WriteString(fmt.Sprintf("(%s) Tj\nT*\n", escapePDFString(line)))
	}
	buf.WriteString("ET\n")
	return buf.Bytes()
}

func assembleSinglePagePDF(content []byte) ([]byte, error) {
	var doc bytes.Buffer
	doc.WriteString("%PDF-1.4\n")

	objects := make([][]byte, 0, 5)
	objects = append(objects, []byte("<< /Type /Catalog /Pages 2 0 R >>\n"))
	objects = append(objects, []byte("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n"))
	pageObj := fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %0.2f %0.2f] /Contents 4 0 R /Resources << /Font << %s 5 0 R >> >> >>\n", pageWidth, pageHeight, defaultFontRef)
	objects = append(objects, []byte(pageObj))
	stream := fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream\n", len(content), content)
	objects = append(objects, []byte(stream))
	fontObj := fmt.Sprintf("<< /Type /Font /Subtype /Type1 /BaseFont /%s >>\n", defaultFont)
	objects = append(objects, []byte(fontObj))

	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = doc.Len()
		doc.WriteString(fmt.Sprintf("%d 0 obj\n", i+1))
		doc.Write(obj)
		doc.WriteString("endobj\n")
	}

	xrefOffset := doc.Len()
	doc.WriteString("xref\n")
	doc.WriteString(fmt.Sprintf("0 %d\n", len(objects)+1))
	doc.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objects); i++ {
		doc.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	doc.WriteString("trailer\n")
	doc.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", len(objects)+1))
	doc.WriteString("startxref\n")
	doc.WriteString(fmt.Sprintf("%d\n", xrefOffset))
	doc.WriteString("%%EOF\n")
	return doc.Bytes(), nil
}

func escapePDFString(in string) string {
	replacer := strings.NewReplacer("(", "\\(", ")", "\\)", "\\", "\\\\")
	return replacer.Replace(in)
}
