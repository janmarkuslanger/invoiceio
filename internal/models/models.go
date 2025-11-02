package models

import "time"

// PaymentDetails capture how the business that issues the invoice expects to be paid.
type PaymentDetails struct {
	BankName     string `json:"bank_name"`
	IBAN         string `json:"iban"`
	BIC          string `json:"bic"`
	PaymentTerms string `json:"payment_terms"`
}

// Profile holds the issuer specific information (who is sending the invoice).
type Profile struct {
	ID             string         `json:"id"`
	DisplayName    string         `json:"display_name"`
	CompanyName    string         `json:"company_name"`
	AddressLine1   string         `json:"address_line_1"`
	AddressLine2   string         `json:"address_line_2"`
	City           string         `json:"city"`
	PostalCode     string         `json:"postal_code"`
	Country        string         `json:"country"`
	Email          string         `json:"email"`
	Phone          string         `json:"phone"`
	TaxID          string         `json:"tax_id"`
	PaymentDetails PaymentDetails `json:"payment_details"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// Customer captures the invoice recipient information.
type Customer struct {
	ID           string    `json:"id"`
	DisplayName  string    `json:"display_name"`
	ContactName  string    `json:"contact_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	AddressLine1 string    `json:"address_line_1"`
	AddressLine2 string    `json:"address_line_2"`
	City         string    `json:"city"`
	PostalCode   string    `json:"postal_code"`
	Country      string    `json:"country"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// InvoiceItem describes an individual line item on an invoice.
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	LineTotal   float64 `json:"line_total"`
}

// Invoice represents an invoice issued to a customer.
type Invoice struct {
	ID             string        `json:"id"`
	Number         string        `json:"number"`
	ProfileID      string        `json:"profile_id"`
	CustomerID     string        `json:"customer_id"`
	IssueDate      time.Time     `json:"issue_date"`
	DueDate        time.Time     `json:"due_date"`
	Items          []InvoiceItem `json:"items"`
	Notes          string        `json:"notes"`
	TaxRatePercent float64       `json:"tax_rate_percent"`
	Subtotal       float64       `json:"subtotal"`
	TaxAmount      float64       `json:"tax_amount"`
	Total          float64       `json:"total"`
	PDFPath        string        `json:"pdf_path"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}
