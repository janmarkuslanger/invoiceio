package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/janmarkuslanger/invoiceio/internal/id"
	"github.com/janmarkuslanger/invoiceio/internal/models"
	"github.com/janmarkuslanger/invoiceio/internal/pdf"
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

	profileList  *widget.List
	customerList *widget.List
	invoiceList  *widget.List

	profileDisplayName *widget.Entry
	profileCompanyName *widget.Entry
	profileAddress1    *widget.Entry
	profileAddress2    *widget.Entry
	profileCity        *widget.Entry
	profilePostalCode  *widget.Entry
	profileCountry     *widget.Entry
	profileEmail       *widget.Entry
	profilePhone       *widget.Entry
	profileTaxID       *widget.Entry
	profileBankName    *widget.Entry
	profileIBAN        *widget.Entry
	profileBIC         *widget.Entry
	profilePayment     *widget.Entry
	profileForm        *widget.Form
	profileEditingID   string
	selectedProfile    int

	customerDisplayName *widget.Entry
	customerContactName *widget.Entry
	customerEmail       *widget.Entry
	customerPhone       *widget.Entry
	customerAddress1    *widget.Entry
	customerAddress2    *widget.Entry
	customerCity        *widget.Entry
	customerPostalCode  *widget.Entry
	customerCountry     *widget.Entry
	customerNotes       *widget.Entry
	customerForm        *widget.Form
	customerEditingID   string
	selectedCustomer    int

	invoiceProfileSelect  *widget.Select
	invoiceCustomerSelect *widget.Select
	invoiceIssueDate      *widget.Entry
	invoiceDueDate        *widget.Entry
	invoiceTaxRate        *widget.Entry
	invoiceNotes          *widget.Entry
	invoiceItems          []models.InvoiceItem
	invoiceItemsList      *widget.List
	invoiceTotalsLabel    *widget.Label
	invoiceDetails        *widget.Label
	invoiceForm           *widget.Form
	invoiceEditingID      string
	selectedInvoice       int
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

	u.refreshProfiles()
	u.refreshCustomers()
	u.refreshInvoices()

	return tabs
}

func (u *UI) makeProfilesTab() fyne.CanvasObject {
	u.profileDisplayName = widget.NewEntry()
	u.profileCompanyName = widget.NewEntry()
	u.profileAddress1 = widget.NewEntry()
	u.profileAddress2 = widget.NewEntry()
	u.profileCity = widget.NewEntry()
	u.profilePostalCode = widget.NewEntry()
	u.profileCountry = widget.NewEntry()
	u.profileEmail = widget.NewEntry()
	u.profilePhone = widget.NewEntry()
	u.profileTaxID = widget.NewEntry()
	u.profileBankName = widget.NewEntry()
	u.profileIBAN = widget.NewEntry()
	u.profileBIC = widget.NewEntry()
	u.profilePayment = widget.NewEntry()

	u.profileForm = widget.NewForm(
		widget.NewFormItem("Display Name", u.profileDisplayName),
		widget.NewFormItem("Company Name", u.profileCompanyName),
		widget.NewFormItem("Address Line 1", u.profileAddress1),
		widget.NewFormItem("Address Line 2", u.profileAddress2),
		widget.NewFormItem("City", u.profileCity),
		widget.NewFormItem("Postal Code", u.profilePostalCode),
		widget.NewFormItem("Country", u.profileCountry),
		widget.NewFormItem("Email", u.profileEmail),
		widget.NewFormItem("Phone", u.profilePhone),
		widget.NewFormItem("Tax ID", u.profileTaxID),
		widget.NewFormItem("Bank Name", u.profileBankName),
		widget.NewFormItem("IBAN", u.profileIBAN),
		widget.NewFormItem("BIC", u.profileBIC),
		widget.NewFormItem("Payment Terms", u.profilePayment),
	)
	u.profileForm.SubmitText = "Save Profile"
	u.profileForm.CancelText = "Clear"
	u.profileForm.OnSubmit = func() {
		if strings.TrimSpace(u.profileDisplayName.Text) == "" {
			dialog.ShowError(fmt.Errorf("display name is required"), u.win)
			return
		}
		now := time.Now()
		profileID := u.profileEditingID
		if profileID == "" {
			profileID = id.New()
		}
		createdAt := now
		if u.profileEditingID != "" {
			if existing, err := u.store.GetProfile(u.profileEditingID); err == nil {
				createdAt = existing.CreatedAt
			}
		}
		profile := models.Profile{
			ID:           profileID,
			DisplayName:  strings.TrimSpace(u.profileDisplayName.Text),
			CompanyName:  strings.TrimSpace(u.profileCompanyName.Text),
			AddressLine1: strings.TrimSpace(u.profileAddress1.Text),
			AddressLine2: strings.TrimSpace(u.profileAddress2.Text),
			City:         strings.TrimSpace(u.profileCity.Text),
			PostalCode:   strings.TrimSpace(u.profilePostalCode.Text),
			Country:      strings.TrimSpace(u.profileCountry.Text),
			Email:        strings.TrimSpace(u.profileEmail.Text),
			Phone:        strings.TrimSpace(u.profilePhone.Text),
			TaxID:        strings.TrimSpace(u.profileTaxID.Text),
			PaymentDetails: models.PaymentDetails{
				BankName:     strings.TrimSpace(u.profileBankName.Text),
				IBAN:         strings.TrimSpace(u.profileIBAN.Text),
				BIC:          strings.TrimSpace(u.profileBIC.Text),
				PaymentTerms: strings.TrimSpace(u.profilePayment.Text),
			},
			CreatedAt: createdAt,
			UpdatedAt: now,
		}
		if err := u.store.SaveProfile(profile); err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		if u.profileEditingID != "" {
			dialog.ShowInformation("Profile updated", fmt.Sprintf("Profile %s updated.", profile.DisplayName), u.win)
		} else {
			dialog.ShowInformation("Profile saved", fmt.Sprintf("Profile %s stored.", profile.DisplayName), u.win)
		}
		u.resetProfileForm()
		u.refreshProfiles()
	}
	u.profileForm.OnCancel = func() {
		u.resetProfileForm()
	}

	u.profileList = widget.NewList(
		func() int { return len(u.profiles) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.profiles) {
				return
			}
			p := u.profiles[id]
			obj.(*widget.Label).SetText(fmt.Sprintf("%s (%s) – %s, %s %s", p.DisplayName, p.CompanyName, p.AddressLine1, p.PostalCode, p.City))
		},
	)
	u.profileList.OnSelected = func(id widget.ListItemID) {
		u.selectedProfile = id
	}

	editButton := widget.NewButton("Edit Selected", func() {
		if u.selectedProfile < 0 || u.selectedProfile >= len(u.profiles) {
			dialog.ShowInformation("Select profile", "Please choose a profile in the list first.", u.win)
			return
		}
		u.loadProfileForEdit(u.profiles[u.selectedProfile])
	})

	actionBar := container.NewHBox(editButton)

	return container.NewBorder(u.profileForm, actionBar, nil, nil, container.NewVScroll(u.profileList))
}

func (u *UI) resetProfileForm() {
	u.profileEditingID = ""
	if u.profileDisplayName != nil {
		u.profileDisplayName.SetText("")
	}
	u.profileCompanyName.SetText("")
	u.profileAddress1.SetText("")
	u.profileAddress2.SetText("")
	u.profileCity.SetText("")
	u.profilePostalCode.SetText("")
	u.profileCountry.SetText("")
	u.profileEmail.SetText("")
	u.profilePhone.SetText("")
	u.profileTaxID.SetText("")
	u.profileBankName.SetText("")
	u.profileIBAN.SetText("")
	u.profileBIC.SetText("")
	u.profilePayment.SetText("")
	if u.profileForm != nil {
		u.profileForm.SubmitText = "Save Profile"
		u.profileForm.Refresh()
	}
}

func (u *UI) loadProfileForEdit(p models.Profile) {
	u.profileEditingID = p.ID
	u.profileDisplayName.SetText(p.DisplayName)
	u.profileCompanyName.SetText(p.CompanyName)
	u.profileAddress1.SetText(p.AddressLine1)
	u.profileAddress2.SetText(p.AddressLine2)
	u.profileCity.SetText(p.City)
	u.profilePostalCode.SetText(p.PostalCode)
	u.profileCountry.SetText(p.Country)
	u.profileEmail.SetText(p.Email)
	u.profilePhone.SetText(p.Phone)
	u.profileTaxID.SetText(p.TaxID)
	u.profileBankName.SetText(p.PaymentDetails.BankName)
	u.profileIBAN.SetText(p.PaymentDetails.IBAN)
	u.profileBIC.SetText(p.PaymentDetails.BIC)
	u.profilePayment.SetText(p.PaymentDetails.PaymentTerms)
	if u.profileForm != nil {
		u.profileForm.SubmitText = "Update Profile"
		u.profileForm.Refresh()
	}
}

func (u *UI) makeCustomersTab() fyne.CanvasObject {
	u.customerDisplayName = widget.NewEntry()
	u.customerContactName = widget.NewEntry()
	u.customerEmail = widget.NewEntry()
	u.customerPhone = widget.NewEntry()
	u.customerAddress1 = widget.NewEntry()
	u.customerAddress2 = widget.NewEntry()
	u.customerCity = widget.NewEntry()
	u.customerPostalCode = widget.NewEntry()
	u.customerCountry = widget.NewEntry()
	u.customerNotes = widget.NewMultiLineEntry()

	u.customerForm = widget.NewForm(
		widget.NewFormItem("Display Name", u.customerDisplayName),
		widget.NewFormItem("Contact Name", u.customerContactName),
		widget.NewFormItem("Email", u.customerEmail),
		widget.NewFormItem("Phone", u.customerPhone),
		widget.NewFormItem("Address Line 1", u.customerAddress1),
		widget.NewFormItem("Address Line 2", u.customerAddress2),
		widget.NewFormItem("City", u.customerCity),
		widget.NewFormItem("Postal Code", u.customerPostalCode),
		widget.NewFormItem("Country", u.customerCountry),
		widget.NewFormItem("Notes", u.customerNotes),
	)
	u.customerForm.SubmitText = "Save Customer"
	u.customerForm.CancelText = "Clear"
	u.customerForm.OnSubmit = func() {
		if strings.TrimSpace(u.customerDisplayName.Text) == "" {
			dialog.ShowError(fmt.Errorf("display name is required"), u.win)
			return
		}
		now := time.Now()
		customerID := u.customerEditingID
		if customerID == "" {
			customerID = id.New()
		}
		createdAt := now
		if u.customerEditingID != "" {
			if existing, err := u.store.GetCustomer(u.customerEditingID); err == nil {
				createdAt = existing.CreatedAt
			}
		}
		customer := models.Customer{
			ID:           customerID,
			DisplayName:  strings.TrimSpace(u.customerDisplayName.Text),
			ContactName:  strings.TrimSpace(u.customerContactName.Text),
			Email:        strings.TrimSpace(u.customerEmail.Text),
			Phone:        strings.TrimSpace(u.customerPhone.Text),
			AddressLine1: strings.TrimSpace(u.customerAddress1.Text),
			AddressLine2: strings.TrimSpace(u.customerAddress2.Text),
			City:         strings.TrimSpace(u.customerCity.Text),
			PostalCode:   strings.TrimSpace(u.customerPostalCode.Text),
			Country:      strings.TrimSpace(u.customerCountry.Text),
			Notes:        strings.TrimSpace(u.customerNotes.Text),
			CreatedAt:    createdAt,
			UpdatedAt:    now,
		}
		if err := u.store.SaveCustomer(customer); err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		if u.customerEditingID != "" {
			dialog.ShowInformation("Customer updated", fmt.Sprintf("Customer %s updated.", customer.DisplayName), u.win)
		} else {
			dialog.ShowInformation("Customer saved", fmt.Sprintf("Customer %s stored.", customer.DisplayName), u.win)
		}
		u.resetCustomerForm()
		u.refreshCustomers()
	}
	u.customerForm.OnCancel = func() {
		u.resetCustomerForm()
	}

	u.customerList = widget.NewList(
		func() int { return len(u.customers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.customers) {
				return
			}
			c := u.customers[id]
			obj.(*widget.Label).SetText(fmt.Sprintf("%s (%s) – %s, %s %s", c.DisplayName, c.ContactName, c.AddressLine1, c.PostalCode, c.City))
		},
	)
	u.customerList.OnSelected = func(id widget.ListItemID) {
		u.selectedCustomer = id
	}

	editButton := widget.NewButton("Edit Selected", func() {
		if u.selectedCustomer < 0 || u.selectedCustomer >= len(u.customers) {
			dialog.ShowInformation("Select customer", "Please choose a customer in the list first.", u.win)
			return
		}
		u.loadCustomerForEdit(u.customers[u.selectedCustomer])
	})

	actionBar := container.NewHBox(editButton)

	return container.NewBorder(u.customerForm, actionBar, nil, nil, container.NewVScroll(u.customerList))
}

func (u *UI) resetCustomerForm() {
	u.customerEditingID = ""
	u.customerDisplayName.SetText("")
	u.customerContactName.SetText("")
	u.customerEmail.SetText("")
	u.customerPhone.SetText("")
	u.customerAddress1.SetText("")
	u.customerAddress2.SetText("")
	u.customerCity.SetText("")
	u.customerPostalCode.SetText("")
	u.customerCountry.SetText("")
	u.customerNotes.SetText("")
	if u.customerForm != nil {
		u.customerForm.SubmitText = "Save Customer"
		u.customerForm.Refresh()
	}
}

func (u *UI) loadCustomerForEdit(c models.Customer) {
	u.customerEditingID = c.ID
	u.customerDisplayName.SetText(c.DisplayName)
	u.customerContactName.SetText(c.ContactName)
	u.customerEmail.SetText(c.Email)
	u.customerPhone.SetText(c.Phone)
	u.customerAddress1.SetText(c.AddressLine1)
	u.customerAddress2.SetText(c.AddressLine2)
	u.customerCity.SetText(c.City)
	u.customerPostalCode.SetText(c.PostalCode)
	u.customerCountry.SetText(c.Country)
	u.customerNotes.SetText(c.Notes)
	if u.customerForm != nil {
		u.customerForm.SubmitText = "Update Customer"
		u.customerForm.Refresh()
	}
}

func (u *UI) refreshProfiles() {
	profiles, err := u.store.ListProfiles()
	if err != nil {
		dialog.ShowError(fmt.Errorf("load profiles: %w", err), u.win)
		return
	}
	u.profiles = profiles
	if u.profileList != nil {
		u.profileList.Refresh()
	}
	u.selectedProfile = -1
	if u.invoiceProfileSelect != nil {
		labels := make([]string, len(profiles))
		for i, p := range profiles {
			labels[i] = u.profileLabel(p)
		}
		u.invoiceProfileSelect.Options = labels
		if len(labels) > 0 && u.invoiceProfileSelect.Selected == "" {
			u.invoiceProfileSelect.SetSelected(labels[0])
		} else {
			u.invoiceProfileSelect.Refresh()
		}
	}
}

func (u *UI) refreshCustomers() {
	customers, err := u.store.ListCustomers()
	if err != nil {
		dialog.ShowError(fmt.Errorf("load customers: %w", err), u.win)
		return
	}
	u.customers = customers
	if u.customerList != nil {
		u.customerList.Refresh()
	}
	u.selectedCustomer = -1
	if u.invoiceCustomerSelect != nil {
		labels := make([]string, len(customers))
		for i, c := range customers {
			labels[i] = u.customerLabel(c)
		}
		u.invoiceCustomerSelect.Options = labels
		if len(labels) > 0 && u.invoiceCustomerSelect.Selected == "" {
			u.invoiceCustomerSelect.SetSelected(labels[0])
		} else {
			u.invoiceCustomerSelect.Refresh()
		}
	}
}

func (u *UI) refreshInvoices() {
	invoices, err := u.store.ListInvoices()
	if err != nil {
		dialog.ShowError(fmt.Errorf("load invoices: %w", err), u.win)
		return
	}
	u.invoices = invoices
	u.invoiceSummaries = make([]string, len(invoices))
	for i, inv := range invoices {
		customerName := inv.CustomerID
		if cust, err := u.store.GetCustomer(inv.CustomerID); err == nil {
			customerName = cust.DisplayName
		}
		u.invoiceSummaries[i] = fmt.Sprintf("%s – %s – Total %.2f", inv.Number, customerName, inv.Total)
	}
	if u.invoiceList != nil {
		u.invoiceList.Refresh()
	}
	u.selectedInvoice = -1
	if u.invoiceDetails != nil {
		if len(invoices) == 0 {
			u.invoiceDetails.SetText("No invoices stored yet.")
		} else {
			u.invoiceDetails.SetText("Select an invoice to see details.")
		}
	}
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
	for _, p := range u.profiles {
		if u.profileLabel(p) == label {
			return p, true
		}
	}
	return models.Profile{}, false
}

func (u *UI) customerByLabel(label string) (models.Customer, bool) {
	for _, c := range u.customers {
		if u.customerLabel(c) == label {
			return c, true
		}
	}
	return models.Customer{}, false
}
func (u *UI) makeInvoicesTab() fyne.CanvasObject {
	u.invoiceProfileSelect = widget.NewSelect([]string{}, nil)
	u.invoiceCustomerSelect = widget.NewSelect([]string{}, nil)
	u.invoiceIssueDate = widget.NewEntry()
	u.invoiceIssueDate.SetText(time.Now().Format("2006-01-02"))
	u.invoiceDueDate = widget.NewEntry()
	u.invoiceDueDate.SetText(time.Now().AddDate(0, 0, 14).Format("2006-01-02"))
	u.invoiceTaxRate = widget.NewEntry()
	u.invoiceTaxRate.SetText("0")
	u.invoiceNotes = widget.NewMultiLineEntry()
	u.invoiceItems = make([]models.InvoiceItem, 0, 4)

	u.invoiceItemsList = widget.NewList(
		func() int { return len(u.invoiceItems) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.invoiceItems) {
				return
			}
			item := u.invoiceItems[id]
			obj.(*widget.Label).SetText(fmt.Sprintf("%s – Qty %.2f @ %.2f => %.2f", item.Description, item.Quantity, item.UnitPrice, item.LineTotal))
		},
	)

	u.invoiceTotalsLabel = widget.NewLabel("Subtotal: 0.00 | Tax: 0.00 | Total: 0.00")
	u.invoiceTotalsLabel.Wrapping = fyne.TextWrapWord

	addItem := widget.NewButton("Add Item", func() {
		descEntry := widget.NewEntry()
		qtyEntry := widget.NewEntry()
		qtyEntry.SetText("1")
		priceEntry := widget.NewEntry()
		priceEntry.SetText("0")
		d := dialog.NewForm("Add Line Item", "Add", "Cancel", []*widget.FormItem{
			widget.NewFormItem("Description", descEntry),
			widget.NewFormItem("Quantity", qtyEntry),
			widget.NewFormItem("Unit Price", priceEntry),
		}, func(confirm bool) {
			if !confirm {
				return
			}
			desc := strings.TrimSpace(descEntry.Text)
			if desc == "" {
				dialog.ShowError(fmt.Errorf("description is required"), u.win)
				return
			}
			qty, err := strconv.ParseFloat(strings.TrimSpace(qtyEntry.Text), 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid quantity: %w", err), u.win)
				return
			}
			price, err := strconv.ParseFloat(strings.TrimSpace(priceEntry.Text), 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid unit price: %w", err), u.win)
				return
			}
			lineTotal := qty * price
			u.invoiceItems = append(u.invoiceItems, models.InvoiceItem{
				Description: desc,
				Quantity:    qty,
				UnitPrice:   price,
				LineTotal:   lineTotal,
			})
			u.invoiceItemsList.Refresh()
			u.updateInvoiceTotals()
		}, u.win)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
	})

	removeItem := widget.NewButton("Remove Last Item", func() {
		if len(u.invoiceItems) == 0 {
			return
		}
		u.invoiceItems = u.invoiceItems[:len(u.invoiceItems)-1]
		u.invoiceItemsList.Refresh()
		u.updateInvoiceTotals()
	})

	u.invoiceTaxRate.OnChanged = func(string) {
		u.updateInvoiceTotals()
	}

	u.invoiceDetails = widget.NewLabel("Select an invoice to see details.")
	u.invoiceDetails.Wrapping = fyne.TextWrapWord

	u.invoiceList = widget.NewList(
		func() int { return len(u.invoices) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.invoiceSummaries) {
				return
			}
			obj.(*widget.Label).SetText(u.invoiceSummaries[id])
		},
	)

	u.invoiceList.OnSelected = func(id widget.ListItemID) {
		u.selectedInvoice = id
		if id < 0 || id >= len(u.invoices) {
			return
		}
		inv := u.invoices[id]
		customerName := inv.CustomerID
		if cust, err := u.store.GetCustomer(inv.CustomerID); err == nil {
			customerName = cust.DisplayName
		}
		profileName := inv.ProfileID
		if prof, err := u.store.GetProfile(inv.ProfileID); err == nil {
			profileName = prof.DisplayName
		}
		u.invoiceDetails.SetText(fmt.Sprintf("Invoice %s\nIssued: %s\nDue: %s\nProfile: %s\nCustomer: %s\nItems: %d\nSubtotal: %.2f\nTax: %.2f\nTotal: %.2f\nPDF: %s",
			inv.Number,
			inv.IssueDate.Format("2006-01-02"),
			inv.DueDate.Format("2006-01-02"),
			profileName,
			customerName,
			len(inv.Items),
			inv.Subtotal,
			inv.TaxAmount,
			inv.Total,
			inv.PDFPath,
		))
	}

	u.invoiceForm = widget.NewForm(
		widget.NewFormItem("Profile", u.invoiceProfileSelect),
		widget.NewFormItem("Customer", u.invoiceCustomerSelect),
		widget.NewFormItem("Issue Date (YYYY-MM-DD)", u.invoiceIssueDate),
		widget.NewFormItem("Due Date (YYYY-MM-DD)", u.invoiceDueDate),
		widget.NewFormItem("Tax Rate (%)", u.invoiceTaxRate),
		widget.NewFormItem("Notes", u.invoiceNotes),
	)
	u.invoiceForm.SubmitText = "Create Invoice"
	u.invoiceForm.CancelText = "Clear"
	u.invoiceForm.OnSubmit = func() {
		profileLabel := u.invoiceProfileSelect.Selected
		customerLabel := u.invoiceCustomerSelect.Selected
		if profileLabel == "" || customerLabel == "" {
			dialog.ShowError(fmt.Errorf("profile and customer must be selected"), u.win)
			return
		}
		if len(u.invoiceItems) == 0 {
			dialog.ShowError(fmt.Errorf("add at least one line item"), u.win)
			return
		}
		profile, ok := u.profileByLabel(profileLabel)
		if !ok {
			dialog.ShowError(fmt.Errorf("selected profile unavailable"), u.win)
			return
		}
		customer, ok := u.customerByLabel(customerLabel)
		if !ok {
			dialog.ShowError(fmt.Errorf("selected customer unavailable"), u.win)
			return
		}
		issue, err := time.Parse("2006-01-02", strings.TrimSpace(u.invoiceIssueDate.Text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid issue date: %w", err), u.win)
			return
		}
		due, err := time.Parse("2006-01-02", strings.TrimSpace(u.invoiceDueDate.Text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid due date: %w", err), u.win)
			return
		}
		taxPercent := 0.0
		if strings.TrimSpace(u.invoiceTaxRate.Text) != "" {
			taxPercent, err = strconv.ParseFloat(strings.TrimSpace(u.invoiceTaxRate.Text), 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid tax rate: %w", err), u.win)
				return
			}
		}

		subtotal := 0.0
		for _, item := range u.invoiceItems {
			subtotal += item.LineTotal
		}
		taxAmount := subtotal * (taxPercent / 100)
		total := subtotal + taxAmount
		now := time.Now()

		invoiceID := u.invoiceEditingID
		if invoiceID == "" {
			invoiceID = id.New()
		}
		invoiceNumber := ""
		pdfPath := ""
		createdAt := now
		if u.invoiceEditingID != "" {
			if existing, err := u.store.GetInvoice(u.invoiceEditingID); err == nil {
				invoiceNumber = existing.Number
				pdfPath = existing.PDFPath
				createdAt = existing.CreatedAt
			}
		}
		if invoiceNumber == "" {
			invoiceNumber = fmt.Sprintf("INV-%s-%s", issue.Format("20060102"), id.Short())
		}
		if pdfPath == "" {
			pdfPath = filepath.Join(u.store.BaseDir(), "pdf", fmt.Sprintf("%s.pdf", strings.ToLower(invoiceNumber)))
		}

		invoice := models.Invoice{
			ID:             invoiceID,
			Number:         invoiceNumber,
			ProfileID:      profile.ID,
			CustomerID:     customer.ID,
			IssueDate:      issue,
			DueDate:        due,
			Items:          append([]models.InvoiceItem(nil), u.invoiceItems...),
			Notes:          strings.TrimSpace(u.invoiceNotes.Text),
			TaxRatePercent: taxPercent,
			Subtotal:       subtotal,
			TaxAmount:      taxAmount,
			Total:          total,
			PDFPath:        pdfPath,
			CreatedAt:      createdAt,
			UpdatedAt:      now,
		}

		if err := u.store.SaveInvoice(invoice); err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		if err := pdf.CreateInvoicePDF(pdfPath, profile, customer, invoice); err != nil {
			dialog.ShowError(fmt.Errorf("failed to create PDF: %w", err), u.win)
			return
		}

		if u.invoiceEditingID != "" {
			dialog.ShowInformation("Invoice updated", fmt.Sprintf("Invoice %s regenerated.\nPDF stored at %s", invoice.Number, pdfPath), u.win)
		} else {
			dialog.ShowInformation("Invoice created", fmt.Sprintf("Invoice %s generated.\nPDF stored at %s", invoice.Number, pdfPath), u.win)
		}

		u.resetInvoiceForm()
		u.refreshInvoices()
	}
	u.invoiceForm.OnCancel = func() {
		u.resetInvoiceForm()
	}

	itemsControls := container.NewVBox(
		container.NewGridWithColumns(2, addItem, removeItem),
		container.NewVScroll(u.invoiceItemsList),
		u.invoiceTotalsLabel,
	)
	left := container.NewVBox(
		u.invoiceForm,
		widget.NewSeparator(),
		itemsControls,
	)

	editButton := widget.NewButton("Edit Selected", func() {
		if u.selectedInvoice < 0 || u.selectedInvoice >= len(u.invoices) {
			dialog.ShowInformation("Select invoice", "Please choose an invoice from the list first.", u.win)
			return
		}
		u.loadInvoiceForEdit(u.invoices[u.selectedInvoice])
	})

	invoiceListPanel := container.NewBorder(nil, editButton, nil, nil, container.NewVScroll(u.invoiceList))
	right := container.NewBorder(invoiceListPanel, nil, nil, nil, container.NewVScroll(u.invoiceDetails))

	split := container.NewHSplit(left, right)
	split.SetOffset(0.52)
	return split
}

func (u *UI) updateInvoiceTotals() {
	subtotal := 0.0
	for _, item := range u.invoiceItems {
		subtotal += item.LineTotal
	}
	taxVal := 0.0
	if u.invoiceTaxRate != nil {
		if v := strings.TrimSpace(u.invoiceTaxRate.Text); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				taxVal = subtotal * (f / 100)
			}
		}
	}
	total := subtotal + taxVal
	if u.invoiceTotalsLabel != nil {
		u.invoiceTotalsLabel.SetText(fmt.Sprintf("Subtotal: %.2f | Tax: %.2f | Total: %.2f", subtotal, taxVal, total))
	}
}

func (u *UI) resetInvoiceForm() {
	u.invoiceEditingID = ""
	u.invoiceIssueDate.SetText(time.Now().Format("2006-01-02"))
	u.invoiceDueDate.SetText(time.Now().AddDate(0, 0, 14).Format("2006-01-02"))
	u.invoiceTaxRate.SetText("0")
	u.invoiceNotes.SetText("")
	u.invoiceItems = u.invoiceItems[:0]
	if u.invoiceItemsList != nil {
		u.invoiceItemsList.Refresh()
	}
	if u.invoiceProfileSelect != nil {
		u.invoiceProfileSelect.Selected = ""
		u.invoiceProfileSelect.Refresh()
	}
	if u.invoiceCustomerSelect != nil {
		u.invoiceCustomerSelect.Selected = ""
		u.invoiceCustomerSelect.Refresh()
	}
	if u.invoiceForm != nil {
		u.invoiceForm.SubmitText = "Create Invoice"
		u.invoiceForm.Refresh()
	}
	u.updateInvoiceTotals()
}

func (u *UI) loadInvoiceForEdit(inv models.Invoice) {
	u.invoiceEditingID = inv.ID
	u.invoiceItems = append([]models.InvoiceItem(nil), inv.Items...)
	if u.invoiceItemsList != nil {
		u.invoiceItemsList.Refresh()
	}

	u.invoiceIssueDate.SetText(inv.IssueDate.Format("2006-01-02"))
	u.invoiceDueDate.SetText(inv.DueDate.Format("2006-01-02"))
	u.invoiceTaxRate.SetText(fmt.Sprintf("%.2f", inv.TaxRatePercent))
	u.invoiceNotes.SetText(inv.Notes)

	if profile, err := u.store.GetProfile(inv.ProfileID); err == nil {
		label := u.profileLabel(profile)
		if contains(u.invoiceProfileSelect.Options, label) {
			u.invoiceProfileSelect.SetSelected(label)
		}
	}
	if customer, err := u.store.GetCustomer(inv.CustomerID); err == nil {
		label := u.customerLabel(customer)
		if contains(u.invoiceCustomerSelect.Options, label) {
			u.invoiceCustomerSelect.SetSelected(label)
		}
	}

	if u.invoiceForm != nil {
		u.invoiceForm.SubmitText = "Update Invoice"
		u.invoiceForm.Refresh()
	}
	u.updateInvoiceTotals()
}

func contains(vals []string, target string) bool {
	for _, v := range vals {
		if v == target {
			return true
		}
	}
	return false
}
