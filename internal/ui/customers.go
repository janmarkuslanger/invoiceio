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

	"github.com/janmarkuslanger/invoiceio/internal/i18n"
	"github.com/janmarkuslanger/invoiceio/internal/id"
	"github.com/janmarkuslanger/invoiceio/internal/models"
)

func (u *UI) makeCustomersTab() fyne.CanvasObject {
	u.customerDetailText = widget.NewRichTextFromMarkdown(i18n.T("customers.detail.empty"))
	detailCard := widget.NewCard(i18n.T("customers.detail.title"), "", u.customerDetailText)

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

	newButton := widget.NewButtonWithIcon(i18n.T("customers.button.new"), themePlusIcon(), func() {
		u.openCustomerDialog(nil)
	})
	u.customerEditButton = widget.NewButton(i18n.T("button.editSelected"), func() {
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
	title := i18n.T("customers.dialog.newTitle")
	submitLabel := i18n.T("customers.dialog.create")
	var current models.Customer
	if isEdit {
		title = i18n.T("customers.dialog.editTitle")
		submitLabel = i18n.T("customers.dialog.update")
		current = *existing
	}

	displayName := widget.NewEntry()
	displayName.SetPlaceHolder(i18n.T("forms.placeholder.required"))
	displayName.Validator = validation.NewRegexp(`\S+`, i18n.T("customers.error.displayNameRequired"))
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
		widget.NewFormItem(i18n.T("customers.form.displayName"), displayName),
		widget.NewFormItem(i18n.T("customers.form.contactName"), contactName),
		widget.NewFormItem(i18n.T("customers.form.email"), email),
		widget.NewFormItem(i18n.T("customers.form.phone"), phone),
		widget.NewFormItem(i18n.T("customers.form.address1"), address1),
		widget.NewFormItem(i18n.T("customers.form.address2"), address2),
		widget.NewFormItem(i18n.T("customers.form.city"), city),
		widget.NewFormItem(i18n.T("customers.form.postalCode"), postalCode),
		widget.NewFormItem(i18n.T("customers.form.country"), country),
		widget.NewFormItem(i18n.T("customers.form.notes"), notes),
	)

	u.showFormDialog(title, submitLabel, form, func() error {
		if strings.TrimSpace(displayName.Text) == "" {
			return fmt.Errorf("%s", i18n.T("customers.error.displayNameRequired"))
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
			return fmt.Errorf("%s: %w", i18n.T("customers.error.save"), err)
		}
		u.refreshCustomers(customer.ID)
		u.lastCustomerID = customer.ID
		if isEdit {
			dialog.ShowInformation(i18n.T("customers.info.updatedTitle"), i18n.T("customers.info.updatedBody", customer.DisplayName), u.win)
		} else {
			dialog.ShowInformation(i18n.T("customers.info.createdTitle"), i18n.T("customers.info.createdBody", customer.DisplayName), u.win)
		}
		return nil
	})
}

func (u *UI) updateCustomerDetail() {
	if u.customerDetailText == nil {
		return
	}
	if u.selectedCustomer < 0 || u.selectedCustomer >= len(u.customers) {
		u.customerDetailText.ParseMarkdown(i18n.T("customers.detail.empty"))
		return
	}
	c := u.customers[u.selectedCustomer]
	lines := []string{
		i18n.T("customers.detail.displayName", c.DisplayName),
		i18n.T("customers.detail.contact", c.ContactName),
		i18n.T("customers.detail.email", c.Email),
		i18n.T("customers.detail.phone", c.Phone),
		"",
		i18n.T("customers.detail.addressTitle"),
	}
	if line := strings.TrimSpace(c.AddressLine1); line != "" {
		lines = append(lines, line)
	}
	if line := strings.TrimSpace(c.AddressLine2); line != "" {
		lines = append(lines, line)
	}
	lines = append(lines, i18n.T("customers.detail.cityPostal", strings.TrimSpace(c.PostalCode), strings.TrimSpace(c.City)))
	lines = append(lines, i18n.T("customers.detail.country", strings.TrimSpace(c.Country)))
	if strings.TrimSpace(c.Notes) != "" {
		lines = append(lines, "", i18n.T("customers.detail.notesTitle"), c.Notes)
	}
	u.customerDetailText.ParseMarkdown(strings.Join(lines, "\n"))
}
