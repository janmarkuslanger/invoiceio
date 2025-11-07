package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/janmarkuslanger/invoiceio/internal/i18n"
	"github.com/janmarkuslanger/invoiceio/internal/models"
)

func dialogError(win fyne.Window, err error) {
	dialog.ShowError(err, win)
}

func (u *UI) profileLabel(p models.Profile) string {
	name := strings.TrimSpace(p.DisplayName)
	if name == "" {
		name = p.CompanyName
	}
	return fmt.Sprintf("%s (%s)", name, p.CompanyName)
}

func (u *UI) customerLabel(c models.Customer) string {
	return fmt.Sprintf("%s (%s)", c.DisplayName, c.ContactName)
}

func (u *UI) profileByLabel(label string) (models.Profile, bool) {
	for _, profile := range u.profiles {
		if u.profileLabel(profile) == label {
			return profile, true
		}
	}
	return models.Profile{}, false
}

func (u *UI) customerByLabel(label string) (models.Customer, bool) {
	for _, customer := range u.customers {
		if u.customerLabel(customer) == label {
			return customer, true
		}
	}
	return models.Customer{}, false
}

func (u *UI) profileOptions() []string {
	options := make([]string, len(u.profiles))
	for idx, profile := range u.profiles {
		options[idx] = u.profileLabel(profile)
	}
	return options
}

func (u *UI) customerOptions() []string {
	options := make([]string, len(u.customers))
	for idx, customer := range u.customers {
		options[idx] = u.customerLabel(customer)
	}
	return options
}

func (u *UI) profileByID(id string) (models.Profile, bool) {
	for _, profile := range u.profiles {
		if profile.ID == id {
			return profile, true
		}
	}
	return models.Profile{}, false
}

func (u *UI) customerByID(id string) (models.Customer, bool) {
	for _, customer := range u.customers {
		if customer.ID == id {
			return customer, true
		}
	}
	return models.Customer{}, false
}

func invoiceBadge(inv models.Invoice) string {
	if !inv.PaidAt.IsZero() {
		return i18n.T("invoices.badge.paid")
	}
	now := time.Now()
	if inv.DueDate.Before(now) {
		return i18n.T("invoices.badge.overdue")
	}
	if inv.DueDate.Sub(now) <= 72*time.Hour {
		return i18n.T("invoices.badge.dueSoon")
	}
	return i18n.T("invoices.badge.onTrack")
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func makeHeaderLabel(text string) *widget.Label {
	return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

func (u *UI) updateInvoiceActionButtons() {
	if u.invoiceEditButton != nil {
		if u.selectedInvoice >= 0 {
			u.invoiceEditButton.Enable()
		} else {
			u.invoiceEditButton.Disable()
		}
	}
	if u.invoicePayButton == nil {
		return
	}
	if u.selectedInvoice >= 0 && u.selectedInvoice < len(u.invoices) {
		inv := u.invoices[u.selectedInvoice]
		if inv.PaidAt.IsZero() {
			u.invoicePayButton.SetText(i18n.T("invoices.button.markPaid"))
		} else {
			u.invoicePayButton.SetText(i18n.T("invoices.button.markUnpaid"))
		}
		u.invoicePayButton.Enable()
	} else {
		u.invoicePayButton.SetText(i18n.T("invoices.button.markPaid"))
		u.invoicePayButton.Disable()
	}
}
