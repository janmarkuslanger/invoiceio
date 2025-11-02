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

func (u *UI) makeProfilesTab() fyne.CanvasObject {
	u.profileDetailText = widget.NewRichTextFromMarkdown("_Select a profile to view details._")
	detailCard := widget.NewCard("Profile Details", "", u.profileDetailText)

	u.profileList = widget.NewList(
		func() int { return len(u.profiles) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.profiles) {
				return
			}
			p := u.profiles[id]
			obj.(*widget.Label).SetText(fmt.Sprintf("%s – %s", p.DisplayName, strings.TrimSpace(p.CompanyName)))
		},
	)
	u.profileList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(u.profiles) {
			return
		}
		u.selectedProfile = id
		u.updateProfileDetail()
		u.profileEditButton.Enable()
	}

	newButton := widget.NewButtonWithIcon("New Profile", themePlusIcon(), func() {
		u.openProfileDialog(nil)
	})
	u.profileEditButton = widget.NewButton("Edit Selected", func() {
		if u.selectedProfile < 0 || u.selectedProfile >= len(u.profiles) {
			return
		}
		profile := u.profiles[u.selectedProfile]
		u.openProfileDialog(&profile)
	})
	u.profileEditButton.Disable()

	actionBar := container.NewHBox(newButton, u.profileEditButton)

	split := container.NewHSplit(
		container.NewMax(u.profileList),
		container.NewMax(detailCard),
	)
	split.SetOffset(0.34)

	return container.NewBorder(actionBar, nil, nil, nil, split)
}

func (u *UI) openProfileDialog(existing *models.Profile) {
	isEdit := existing != nil
	title := "New Profile"
	submitLabel := "Create"
	var current models.Profile
	if isEdit {
		title = "Edit Profile"
		submitLabel = "Update"
		current = *existing
	}

	displayName := widget.NewEntry()
	displayName.SetPlaceHolder("Required")
	displayName.Validator = validation.NewRegexp(`\S+`, "Display name is required")
	companyName := widget.NewEntry()
	address1 := widget.NewEntry()
	address2 := widget.NewEntry()
	city := widget.NewEntry()
	postalCode := widget.NewEntry()
	country := widget.NewEntry()
	email := widget.NewEntry()
	phone := widget.NewEntry()
	taxID := widget.NewEntry()
	bankName := widget.NewEntry()
	iban := widget.NewEntry()
	bic := widget.NewEntry()
	paymentTerms := widget.NewEntry()

	if isEdit {
		displayName.SetText(current.DisplayName)
		companyName.SetText(current.CompanyName)
		address1.SetText(current.AddressLine1)
		address2.SetText(current.AddressLine2)
		city.SetText(current.City)
		postalCode.SetText(current.PostalCode)
		country.SetText(current.Country)
		email.SetText(current.Email)
		phone.SetText(current.Phone)
		taxID.SetText(current.TaxID)
		bankName.SetText(current.PaymentDetails.BankName)
		iban.SetText(current.PaymentDetails.IBAN)
		bic.SetText(current.PaymentDetails.BIC)
		paymentTerms.SetText(current.PaymentDetails.PaymentTerms)
	}

	form := widget.NewForm(
		widget.NewFormItem("Display Name", displayName),
		widget.NewFormItem("Company Name", companyName),
		widget.NewFormItem("Address Line 1", address1),
		widget.NewFormItem("Address Line 2", address2),
		widget.NewFormItem("City", city),
		widget.NewFormItem("Postal Code", postalCode),
		widget.NewFormItem("Country", country),
		widget.NewFormItem("Email", email),
		widget.NewFormItem("Phone", phone),
		widget.NewFormItem("Tax ID", taxID),
		widget.NewFormItem("Bank Name", bankName),
		widget.NewFormItem("IBAN", iban),
		widget.NewFormItem("BIC", bic),
		widget.NewFormItem("Payment Terms", paymentTerms),
	)

	u.showFormDialog(title, submitLabel, form, func() error {
		if strings.TrimSpace(displayName.Text) == "" {
			return fmt.Errorf("display name is required")
		}
		now := time.Now()
		profileID := ""
		createdAt := now
		if isEdit {
			profileID = current.ID
			createdAt = current.CreatedAt
		} else {
			profileID = id.New()
		}
		profile := models.Profile{
			ID:           profileID,
			DisplayName:  strings.TrimSpace(displayName.Text),
			CompanyName:  strings.TrimSpace(companyName.Text),
			AddressLine1: strings.TrimSpace(address1.Text),
			AddressLine2: strings.TrimSpace(address2.Text),
			City:         strings.TrimSpace(city.Text),
			PostalCode:   strings.TrimSpace(postalCode.Text),
			Country:      strings.TrimSpace(country.Text),
			Email:        strings.TrimSpace(email.Text),
			Phone:        strings.TrimSpace(phone.Text),
			TaxID:        strings.TrimSpace(taxID.Text),
			PaymentDetails: models.PaymentDetails{
				BankName:     strings.TrimSpace(bankName.Text),
				IBAN:         strings.TrimSpace(iban.Text),
				BIC:          strings.TrimSpace(bic.Text),
				PaymentTerms: strings.TrimSpace(paymentTerms.Text),
			},
			CreatedAt: createdAt,
			UpdatedAt: now,
		}
		if err := u.store.SaveProfile(profile); err != nil {
			return fmt.Errorf("save profile: %w", err)
		}
		u.refreshProfiles(profile.ID)
		u.lastProfileID = profile.ID
		if isEdit {
			dialog.ShowInformation("Profile updated", fmt.Sprintf("Profile %s updated.", profile.DisplayName), u.win)
		} else {
			dialog.ShowInformation("Profile created", fmt.Sprintf("Profile %s stored.", profile.DisplayName), u.win)
		}
		return nil
	})
}

func (u *UI) updateProfileDetail() {
	if u.profileDetailText == nil {
		return
	}
	if u.selectedProfile < 0 || u.selectedProfile >= len(u.profiles) {
		u.profileDetailText.ParseMarkdown("_Select a profile to view details._")
		return
	}
	p := u.profiles[u.selectedProfile]
	lines := []string{
		fmt.Sprintf("**Display Name:** %s", p.DisplayName),
		fmt.Sprintf("**Company:** %s", p.CompanyName),
		fmt.Sprintf("**Email:** %s", p.Email),
		fmt.Sprintf("**Phone:** %s", p.Phone),
		fmt.Sprintf("**Tax ID:** %s", p.TaxID),
		"",
		"**Address**",
		strings.TrimSpace(p.AddressLine1),
		strings.TrimSpace(p.AddressLine2),
		fmt.Sprintf("%s %s", strings.TrimSpace(p.PostalCode), strings.TrimSpace(p.City)),
		strings.TrimSpace(p.Country),
		"",
		"**Payment Details**",
		fmt.Sprintf("Bank: %s", p.PaymentDetails.BankName),
		fmt.Sprintf("IBAN: %s", p.PaymentDetails.IBAN),
		fmt.Sprintf("BIC: %s", p.PaymentDetails.BIC),
		fmt.Sprintf("Terms: %s", p.PaymentDetails.PaymentTerms),
	}
	u.profileDetailText.ParseMarkdown(strings.Join(lines, "\n"))
}
