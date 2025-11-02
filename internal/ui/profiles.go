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

func (u *UI) makeProfilesTab() fyne.CanvasObject {
	u.profileDetailText = widget.NewRichTextFromMarkdown(i18n.T("profiles.detail.empty"))
	detailCard := widget.NewCard(i18n.T("profiles.detail.title"), "", u.profileDetailText)

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

	newButton := widget.NewButtonWithIcon(i18n.T("profiles.button.new"), themePlusIcon(), func() {
		u.openProfileDialog(nil)
	})
	u.profileEditButton = widget.NewButton(i18n.T("button.editSelected"), func() {
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
	title := i18n.T("profiles.dialog.newTitle")
	submitLabel := i18n.T("profiles.dialog.create")
	var current models.Profile
	if isEdit {
		title = i18n.T("profiles.dialog.editTitle")
		submitLabel = i18n.T("profiles.dialog.update")
		current = *existing
	}

	displayName := widget.NewEntry()
	displayName.SetPlaceHolder(i18n.T("forms.placeholder.required"))
	displayName.Validator = validation.NewRegexp(`\S+`, i18n.T("profiles.error.displayNameRequired"))
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
		widget.NewFormItem(i18n.T("profiles.form.displayName"), displayName),
		widget.NewFormItem(i18n.T("profiles.form.companyName"), companyName),
		widget.NewFormItem(i18n.T("profiles.form.address1"), address1),
		widget.NewFormItem(i18n.T("profiles.form.address2"), address2),
		widget.NewFormItem(i18n.T("profiles.form.city"), city),
		widget.NewFormItem(i18n.T("profiles.form.postalCode"), postalCode),
		widget.NewFormItem(i18n.T("profiles.form.country"), country),
		widget.NewFormItem(i18n.T("profiles.form.email"), email),
		widget.NewFormItem(i18n.T("profiles.form.phone"), phone),
		widget.NewFormItem(i18n.T("profiles.form.taxID"), taxID),
		widget.NewFormItem(i18n.T("profiles.form.bankName"), bankName),
		widget.NewFormItem(i18n.T("profiles.form.iban"), iban),
		widget.NewFormItem(i18n.T("profiles.form.bic"), bic),
		widget.NewFormItem(i18n.T("profiles.form.paymentTerms"), paymentTerms),
	)

	u.showFormDialog(title, submitLabel, form, func() error {
		if strings.TrimSpace(displayName.Text) == "" {
			return fmt.Errorf("%s", i18n.T("profiles.error.displayNameRequired"))
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
			return fmt.Errorf("%s: %w", i18n.T("profiles.error.save"), err)
		}
		u.refreshProfiles(profile.ID)
		u.lastProfileID = profile.ID
		if isEdit {
			dialog.ShowInformation(i18n.T("profiles.info.updatedTitle"), i18n.T("profiles.info.updatedBody", profile.DisplayName), u.win)
		} else {
			dialog.ShowInformation(i18n.T("profiles.info.createdTitle"), i18n.T("profiles.info.createdBody", profile.DisplayName), u.win)
		}
		return nil
	})
}

func (u *UI) updateProfileDetail() {
	if u.profileDetailText == nil {
		return
	}
	if u.selectedProfile < 0 || u.selectedProfile >= len(u.profiles) {
		u.profileDetailText.ParseMarkdown(i18n.T("profiles.detail.empty"))
		return
	}
	p := u.profiles[u.selectedProfile]
	lines := []string{
		i18n.T("profiles.detail.displayName", p.DisplayName),
		i18n.T("profiles.detail.company", p.CompanyName),
		i18n.T("profiles.detail.email", p.Email),
		i18n.T("profiles.detail.phone", p.Phone),
		i18n.T("profiles.detail.taxID", p.TaxID),
	}
	lines = append(lines, "")
	lines = append(lines, i18n.T("profiles.detail.addressTitle"))
	if line := strings.TrimSpace(p.AddressLine1); line != "" {
		lines = append(lines, line)
	}
	if line := strings.TrimSpace(p.AddressLine2); line != "" {
		lines = append(lines, line)
	}
	lines = append(lines, i18n.T("profiles.detail.cityPostal", strings.TrimSpace(p.PostalCode), strings.TrimSpace(p.City)))
	lines = append(lines, i18n.T("profiles.detail.country", strings.TrimSpace(p.Country)))
	lines = append(lines, "")
	lines = append(lines, i18n.T("profiles.detail.paymentTitle"))
	if val := strings.TrimSpace(p.PaymentDetails.BankName); val != "" {
		lines = append(lines, i18n.T("profiles.detail.bank", val))
	}
	if val := strings.TrimSpace(p.PaymentDetails.IBAN); val != "" {
		lines = append(lines, i18n.T("profiles.detail.iban", val))
	}
	if val := strings.TrimSpace(p.PaymentDetails.BIC); val != "" {
		lines = append(lines, i18n.T("profiles.detail.bic", val))
	}
	if val := strings.TrimSpace(p.PaymentDetails.PaymentTerms); val != "" {
		lines = append(lines, i18n.T("profiles.detail.paymentTerms", val))
	}
	u.profileDetailText.ParseMarkdown(strings.Join(lines, "\n"))
}
