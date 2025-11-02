package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/janmarkuslanger/invoiceio/internal/i18n"
)

func (u *UI) showFormDialog(title, submitLabel string, form *widget.Form, submit func() error) {
	status := widget.NewLabel("")
	status.Wrapping = fyne.TextWrapWord
	status.Hide()

	save := widget.NewButton(submitLabel, nil)
	cancel := widget.NewButton(i18n.T("common.cancel"), nil)
	buttons := container.NewHBox(layout.NewSpacer(), cancel, save)
	content := container.NewVBox(form, status, buttons)

	dlg := dialog.NewCustomWithoutButtons(title, content, u.win)

	cancel.OnTapped = func() {
		dlg.Hide()
	}

	save.OnTapped = func() {
		if err := submit(); err != nil {
			status.SetText(err.Error())
			status.Show()
			return
		}
		status.Hide()
		dlg.Hide()
	}

	dlg.Resize(fyne.NewSize(520, form.MinSize().Height+120))
	dlg.Show()
}
