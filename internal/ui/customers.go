package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/janmarkuslanger/invoiceio/internal/id"
	"github.com/janmarkuslanger/invoiceio/internal/models"
)

func (u *UI) makeCustomersTab() fyne.CanvasObject {
	u.customerDetailText = widget.NewRichTextFromMarkdown("_Select a customer to view details._")
	detailCard := widget.NewCard("Customer Details", "", u.customerDetailText)

	u.customerList = widget.NewList(
		func() int { return len(u.customers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.customers) {
				return
			}
			c := u.customers[id]
			label := fmt.Sprintf("%s – %s", c.DisplayName, strings.TrimSpace(c.ContactName))
			obj.(*widget.Label).SetText(label)
		},
	)
	u.customerList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(u.customers) {
			return
		}
		u.selectedCustomer = id
		u.updateCustomerDetail()
		u.customerEditButton.Enable()
	}

	newButton := widget.NewButtonWithIcon("New Customer", themePlusIcon(), func() {
		u.openCustomerDialog(nil)
	})
	u.customerEditButton = widget.NewButton("Edit Selected", func() {
		if u.selectedCustomer < 0 || u.selectedCustomer >= len(u.customers) {
			return
		}
		customer := u.customers[u.selectedCustomer]
		u.openCustomerDialog(&customer)
	})
	u.customerEditButton.Disable()

	actionBar := container.NewHBox(newButton, u.customerEditButton)

	split := container.NewHSplit(
		container.NewMax(u.customerList),
		container.NewMax(detailCard),
	)
	split.SetOffset(0.34)

	return container.NewBorder(actionBar, nil, nil, nil, split)
}

func (u *UI) openCustomerDialog(existing *models.Customer) {
	isEdit := existing != nil
	title := "New Customer"
	submitLabel := "Create"
	var current models.Customer
	if isEdit {
		title = "Edit Customer"
		submitLabel = "Update"
		current = *existing
	}

	displayName := widget.NewEntry()
	displayName.SetPlaceHolder("Required")
	displayName.Validator = validation.NewRegexp(`\S+`, "Display name is required")
	contactName := widget.NewEntry()
	email := widget.NewEntry()
	phone := widget.NewEntry()
	address1 := widget.NewEntry()
	address2 := widget.NewEntry()
	city := widget.NewEntry()
	postalCode := widget.NewEntry()
	country := widget.NewEntry()
	notes := widget.NewMultiLineEntry()

	if isEdit {
		displayName.SetText(current.DisplayName)
		contactName.SetText(current.ContactName)
		email.SetText(current.Email)
		phone.SetText(current.Phone)
		address1.SetText(current.AddressLine1)
		address2.SetText(current.AddressLine2)
		city.SetText(current.City)
		postalCode.SetText(current.PostalCode)
		country.SetText(current.Country)
		notes.SetText(current.Notes)
	}

	form := widget.NewForm(
		widget.NewFormItem("Display Name", displayName),
		widget.NewFormItem("Contact Name", contactName),
		widget.NewFormItem("Email", email),
		widget.NewFormItem("Phone", phone),
		widget.NewFormItem("Address Line 1", address1),
		widget.NewFormItem("Address Line 2", address2),
		widget.NewFormItem("City", city),
		widget.NewFormItem("Postal Code", postalCode),
		widget.NewFormItem("Country", country),
		widget.NewFormItem("Notes", notes),
	)

	u.showFormDialog(title, submitLabel, form, func() error {
		if strings.TrimSpace(displayName.Text) == "" {
			return fmt.Errorf("display name is required")
		}
		now := time.Now()
		customerID := ""
		createdAt := now
		if isEdit {
			customerID = current.ID
			createdAt = current.CreatedAt
		} else {
			customerID = id.New()
		}
		customer := models.Customer{
			ID:           customerID,
			DisplayName:  strings.TrimSpace(displayName.Text),
			ContactName:  strings.TrimSpace(contactName.Text),
			Email:        strings.TrimSpace(email.Text),
			Phone:        strings.TrimSpace(phone.Text),
			AddressLine1: strings.TrimSpace(address1.Text),
			AddressLine2: strings.TrimSpace(address2.Text),
			City:         strings.TrimSpace(city.Text),
			PostalCode:   strings.TrimSpace(postalCode.Text),
			Country:      strings.TrimSpace(country.Text),
			Notes:        strings.TrimSpace(notes.Text),
			CreatedAt:    createdAt,
			UpdatedAt:    now,
		}
		if err := u.store.SaveCustomer(customer); err != nil {
			return fmt.Errorf("save customer: %w", err)
		}
		u.refreshCustomers(customer.ID)
		u.lastCustomerID = customer.ID
		if isEdit {
			dialog.ShowInformation("Customer updated", fmt.Sprintf("Customer %s updated.", customer.DisplayName), u.win)
		} else {
			dialog.ShowInformation("Customer created", fmt.Sprintf("Customer %s stored.", customer.DisplayName), u.win)
		}
		return nil
	})
}

func (u *UI) updateCustomerDetail() {
	if u.customerDetailText == nil {
		return
	}
	if u.selectedCustomer < 0 || u.selectedCustomer >= len(u.customers) {
		u.customerDetailText.ParseMarkdown("_Select a customer to view details._")
		return
	}
	c := u.customers[u.selectedCustomer]
	lines := []string{
		fmt.Sprintf("**Display Name:** %s", c.DisplayName),
		fmt.Sprintf("**Contact:** %s", c.ContactName),
		fmt.Sprintf("**Email:** %s", c.Email),
		fmt.Sprintf("**Phone:** %s", c.Phone),
		"",
		"**Address**",
		strings.TrimSpace(c.AddressLine1),
		strings.TrimSpace(c.AddressLine2),
		fmt.Sprintf("%s %s", strings.TrimSpace(c.PostalCode), strings.TrimSpace(c.City)),
		strings.TrimSpace(c.Country),
	}
	if strings.TrimSpace(c.Notes) != "" {
		lines = append(lines, "", "**Notes**", c.Notes)
	}
	u.customerDetailText.ParseMarkdown(strings.Join(lines, "\n"))
}
