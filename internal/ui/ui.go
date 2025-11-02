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

	invoiceProfileSelect  *widget.Select
	invoiceCustomerSelect *widget.Select
	invoiceDetails        *widget.Label
}

// New initialises a UI helper bound to the given storage and window.
func New(store *storage.Storage, win fyne.Window) *UI {
	return &UI{
		store: store,
		win:   win,
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
	displayName := widget.NewEntry()
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
	form.SubmitText = "Save Profile"
	form.CancelText = "Clear"
	form.OnSubmit = func() {
		if strings.TrimSpace(displayName.Text) == "" {
			dialog.ShowError(fmt.Errorf("display name is required"), u.win)
			return
		}
		now := time.Now()
		profile := models.Profile{
			ID:           id.New(),
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
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := u.store.SaveProfile(profile); err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		dialog.ShowInformation("Profile saved", fmt.Sprintf("Profile %s stored.", profile.DisplayName), u.win)
		displayName.SetText("")
		companyName.SetText("")
		address1.SetText("")
		address2.SetText("")
		city.SetText("")
		postalCode.SetText("")
		country.SetText("")
		email.SetText("")
		phone.SetText("")
		taxID.SetText("")
		bankName.SetText("")
		iban.SetText("")
		bic.SetText("")
		paymentTerms.SetText("")
		u.refreshProfiles()
	}
	form.OnCancel = func() {
		displayName.SetText("")
		companyName.SetText("")
		address1.SetText("")
		address2.SetText("")
		city.SetText("")
		postalCode.SetText("")
		country.SetText("")
		email.SetText("")
		phone.SetText("")
		taxID.SetText("")
		bankName.SetText("")
		iban.SetText("")
		bic.SetText("")
		paymentTerms.SetText("")
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

	return container.NewBorder(form, nil, nil, nil, container.NewVScroll(u.profileList))
}

func (u *UI) makeCustomersTab() fyne.CanvasObject {
	displayName := widget.NewEntry()
	contactName := widget.NewEntry()
	email := widget.NewEntry()
	phone := widget.NewEntry()
	address1 := widget.NewEntry()
	address2 := widget.NewEntry()
	city := widget.NewEntry()
	postalCode := widget.NewEntry()
	country := widget.NewEntry()
	notes := widget.NewMultiLineEntry()

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
	form.SubmitText = "Save Customer"
	form.CancelText = "Clear"
	form.OnSubmit = func() {
		if strings.TrimSpace(displayName.Text) == "" {
			dialog.ShowError(fmt.Errorf("display name is required"), u.win)
			return
		}
		now := time.Now()
		customer := models.Customer{
			ID:           id.New(),
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
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := u.store.SaveCustomer(customer); err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		dialog.ShowInformation("Customer saved", fmt.Sprintf("Customer %s stored.", customer.DisplayName), u.win)
		displayName.SetText("")
		contactName.SetText("")
		email.SetText("")
		phone.SetText("")
		address1.SetText("")
		address2.SetText("")
		city.SetText("")
		postalCode.SetText("")
		country.SetText("")
		notes.SetText("")
		u.refreshCustomers()
	}
	form.OnCancel = func() {
		displayName.SetText("")
		contactName.SetText("")
		email.SetText("")
		phone.SetText("")
		address1.SetText("")
		address2.SetText("")
		city.SetText("")
		postalCode.SetText("")
		country.SetText("")
		notes.SetText("")
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

	return container.NewBorder(form, nil, nil, nil, container.NewVScroll(u.customerList))
}

func (u *UI) makeInvoicesTab() fyne.CanvasObject {
	u.invoiceProfileSelect = widget.NewSelect([]string{}, nil)
	u.invoiceCustomerSelect = widget.NewSelect([]string{}, nil)
	issueDate := widget.NewEntry()
	issueDate.SetText(time.Now().Format("2006-01-02"))
	dueDefault := time.Now().AddDate(0, 0, 14)
	dueDate := widget.NewEntry()
	dueDate.SetText(dueDefault.Format("2006-01-02"))
	taxRate := widget.NewEntry()
	taxRate.SetText("0")
	notes := widget.NewMultiLineEntry()

	items := make([]models.InvoiceItem, 0)
	itemsList := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(items) {
				return
			}
			item := items[id]
			obj.(*widget.Label).SetText(fmt.Sprintf("%s – Qty %.2f @ %.2f => %.2f", item.Description, item.Quantity, item.UnitPrice, item.LineTotal))
		},
	)
	totalsLabel := widget.NewLabel("Subtotal: 0.00 | Tax: 0.00 | Total: 0.00")
	totalsLabel.Wrapping = fyne.TextWrapWord

	updateTotals := func() {
		subtotal := 0.0
		for _, itm := range items {
			subtotal += itm.LineTotal
		}
		taxVal := 0.0
		if v := strings.TrimSpace(taxRate.Text); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				taxVal = subtotal * (f / 100)
			}
		}
		total := subtotal + taxVal
		totalsLabel.SetText(fmt.Sprintf("Subtotal: %.2f | Tax: %.2f | Total: %.2f", subtotal, taxVal, total))
	}

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
			items = append(items, models.InvoiceItem{
				Description: desc,
				Quantity:    qty,
				UnitPrice:   price,
				LineTotal:   lineTotal,
			})
			itemsList.Refresh()
			updateTotals()
		}, u.win)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
	})

	removeItem := widget.NewButton("Remove Last Item", func() {
		if len(items) == 0 {
			return
		}
		items = items[:len(items)-1]
		itemsList.Refresh()
		updateTotals()
	})

	taxRate.OnChanged = func(string) {
		updateTotals()
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

	form := widget.NewForm(
		widget.NewFormItem("Profile", u.invoiceProfileSelect),
		widget.NewFormItem("Customer", u.invoiceCustomerSelect),
		widget.NewFormItem("Issue Date (YYYY-MM-DD)", issueDate),
		widget.NewFormItem("Due Date (YYYY-MM-DD)", dueDate),
		widget.NewFormItem("Tax Rate (%)", taxRate),
		widget.NewFormItem("Notes", notes),
	)
	form.SubmitText = "Create Invoice"
	form.CancelText = "Clear"
	form.OnSubmit = func() {
		profileLabel := u.invoiceProfileSelect.Selected
		customerLabel := u.invoiceCustomerSelect.Selected
		if profileLabel == "" || customerLabel == "" {
			dialog.ShowError(fmt.Errorf("profile and customer must be selected"), u.win)
			return
		}
		if len(items) == 0 {
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
		issue, err := time.Parse("2006-01-02", strings.TrimSpace(issueDate.Text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid issue date: %w", err), u.win)
			return
		}
		due, err := time.Parse("2006-01-02", strings.TrimSpace(dueDate.Text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid due date: %w", err), u.win)
			return
		}
		taxPercent := 0.0
		if strings.TrimSpace(taxRate.Text) != "" {
			taxPercent, err = strconv.ParseFloat(strings.TrimSpace(taxRate.Text), 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid tax rate: %w", err), u.win)
				return
			}
		}

		subtotal := 0.0
		for _, item := range items {
			subtotal += item.LineTotal
		}
		taxAmount := subtotal * (taxPercent / 100)
		total := subtotal + taxAmount
		now := time.Now()

		invoiceNumber := fmt.Sprintf("INV-%s-%s", issue.Format("20060102"), id.Short())
		pdfDir := filepath.Join(u.store.BaseDir(), "pdf")
		pdfPath := filepath.Join(pdfDir, fmt.Sprintf("%s.pdf", strings.ToLower(invoiceNumber)))

		invoice := models.Invoice{
			ID:             id.New(),
			Number:         invoiceNumber,
			ProfileID:      profile.ID,
			CustomerID:     customer.ID,
			IssueDate:      issue,
			DueDate:        due,
			Items:          append([]models.InvoiceItem(nil), items...),
			Notes:          strings.TrimSpace(notes.Text),
			TaxRatePercent: taxPercent,
			Subtotal:       subtotal,
			TaxAmount:      taxAmount,
			Total:          total,
			PDFPath:        pdfPath,
			CreatedAt:      now,
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

		dialog.ShowInformation("Invoice created", fmt.Sprintf("Invoice %s generated.\nPDF stored at %s", invoice.Number, pdfPath), u.win)

		items = items[:0]
		itemsList.Refresh()
		updateTotals()
		notes.SetText("")
		u.refreshInvoices()
	}
	form.OnCancel = func() {
		items = items[:0]
		itemsList.Refresh()
		updateTotals()
		issueDate.SetText(time.Now().Format("2006-01-02"))
		dueDate.SetText(time.Now().AddDate(0, 0, 14).Format("2006-01-02"))
		taxRate.SetText("0")
		notes.SetText("")
	}

	itemButtons := container.NewGridWithColumns(2, addItem, removeItem)
	itemsSection := container.NewVBox(
		itemButtons,
		container.NewVScroll(itemsList),
		totalsLabel,
	)
	left := container.NewVBox(
		form,
		widget.NewSeparator(),
		itemsSection,
	)

	invoiceListPanel := container.NewBorder(nil, nil, nil, nil, container.NewVScroll(u.invoiceList))
	right := container.NewBorder(invoiceListPanel, nil, nil, nil, container.NewVScroll(u.invoiceDetails))

	split := container.NewHSplit(left, right)
	split.SetOffset(0.52)
	return split
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
	if len(invoices) == 0 && u.invoiceDetails != nil {
		u.invoiceDetails.SetText("No invoices stored yet.")
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
