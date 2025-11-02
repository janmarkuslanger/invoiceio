package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/janmarkuslanger/invoiceio/internal/i18n"
	"github.com/janmarkuslanger/invoiceio/internal/models"
	"github.com/janmarkuslanger/invoiceio/internal/storage"
)

// UI coordinates the desktop experience using Fyne widgets.
type UI struct {
	store *storage.Storage
	win   fyne.Window

	profiles         []models.Profile
	customers        []models.Customer
	invoices         []models.Invoice
	invoiceSummaries []string

	profileList       *widget.List
	profileDetailText *widget.RichText
	profileEditButton *widget.Button
	selectedProfile   int

	customerList       *widget.List
	customerDetailText *widget.RichText
	customerEditButton *widget.Button
	selectedCustomer   int

	invoiceList       *widget.List
	invoiceDetailText *widget.RichText
	invoiceEditButton *widget.Button
	selectedInvoice   int

	lastProfileID  string
	lastCustomerID string
}

// New initialises a UI helper bound to the given storage and window.
func New(store *storage.Storage, win fyne.Window) *UI {
	return &UI{
		store:            store,
		win:              win,
		selectedProfile:  -1,
		selectedCustomer: -1,
		selectedInvoice:  -1,
	}
}

// Build assembles the application layout.
func (u *UI) Build() fyne.CanvasObject {
	profilesTab := container.NewTabItem(i18n.T("tabs.profiles"), u.makeProfilesTab())
	customersTab := container.NewTabItem(i18n.T("tabs.customers"), u.makeCustomersTab())
	invoicesTab := container.NewTabItem(i18n.T("tabs.invoices"), u.makeInvoicesTab())

	tabs := container.NewAppTabs(profilesTab, customersTab, invoicesTab)
	tabs.SetTabLocation(container.TabLocationTop)

	newProfileButton := widget.NewButtonWithIcon(i18n.T("toolbar.newProfile"), theme.AccountIcon(), func() {
		u.openProfileDialog(nil)
	})
	newCustomerButton := widget.NewButtonWithIcon(i18n.T("toolbar.newCustomer"), theme.AccountIcon(), func() {
		u.openCustomerDialog(nil)
	})
	newInvoiceButton := widget.NewButtonWithIcon(i18n.T("toolbar.newInvoice"), theme.DocumentCreateIcon(), func() {
		if len(u.profiles) == 0 || len(u.customers) == 0 {
			dialog.ShowInformation(i18n.T("messages.setupRequired.title"), i18n.T("messages.setupRequired.body"), u.win)
			return
		}
		u.openInvoiceDialog(nil)
	})
	locales := i18n.Supported()
	labels := make([]string, len(locales))
	labelToLocale := make(map[string]i18n.Locale, len(locales))
	for i, loc := range locales {
		label := i18n.DisplayName(loc)
		labels[i] = label
		labelToLocale[label] = loc
	}
	currentLabel := i18n.DisplayName(i18n.Current())
	localeSelect := widget.NewSelect(labels, func(label string) {
		loc, ok := labelToLocale[label]
		if !ok || loc == i18n.Current() {
			return
		}
		if err := i18n.SetLocale(loc); err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		u.win.SetContent(u.Build())
	})
	localeSelect.SetSelected(currentLabel)

	languageLabel := widget.NewLabel(i18n.T("toolbar.language"))
	toolbar := container.NewHBox(newProfileButton, newCustomerButton, newInvoiceButton, languageLabel, localeSelect, layout.NewSpacer())
	top := container.NewVBox(toolbar, widget.NewSeparator())

	u.refreshProfiles()
	u.refreshCustomers()
	u.refreshInvoices()

	return container.NewBorder(top, nil, nil, nil, tabs)
}

func themePlusIcon() fyne.Resource {
	return theme.ContentAddIcon()
}
