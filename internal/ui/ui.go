package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

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
	profilesTab := container.NewTabItem("Profiles", u.makeProfilesTab())
	customersTab := container.NewTabItem("Customers", u.makeCustomersTab())
	invoicesTab := container.NewTabItem("Invoices", u.makeInvoicesTab())

	tabs := container.NewAppTabs(profilesTab, customersTab, invoicesTab)
	tabs.SetTabLocation(container.TabLocationTop)

	newProfileButton := widget.NewButtonWithIcon("New Profile", theme.AccountIcon(), func() {
		u.openProfileDialog(nil)
	})
	newCustomerButton := widget.NewButtonWithIcon("New Customer", theme.AccountIcon(), func() {
		u.openCustomerDialog(nil)
	})
	newInvoiceButton := widget.NewButtonWithIcon("New Invoice", theme.DocumentCreateIcon(), func() {
		if len(u.profiles) == 0 || len(u.customers) == 0 {
			dialog.ShowInformation("Setup required", "Create at least one profile and one customer first.", u.win)
			return
		}
		u.openInvoiceDialog(nil)
	})
	toolbar := container.NewHBox(newProfileButton, newCustomerButton, newInvoiceButton, layout.NewSpacer())
	top := container.NewVBox(toolbar, widget.NewSeparator())

	u.refreshProfiles()
	u.refreshCustomers()
	u.refreshInvoices()

	return container.NewBorder(top, nil, nil, nil, tabs)
}

func themePlusIcon() fyne.Resource {
	return theme.ContentAddIcon()
}
