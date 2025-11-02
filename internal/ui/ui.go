package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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

	content := container.NewBorder(actionBar, nil, nil, nil, split)
	return content
}

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

	content := container.NewBorder(actionBar, nil, nil, nil, split)
	return content
}

func (u *UI) makeInvoicesTab() fyne.CanvasObject {
	u.invoiceDetailText = widget.NewRichTextFromMarkdown("_Select an invoice to view details._")
	detailCard := widget.NewCard("Invoice Details", "", u.invoiceDetailText)

	u.invoiceList = widget.NewList(
		func() int { return len(u.invoiceSummaries) },
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
		u.selectedInvoice = id
		u.updateInvoiceDetail()
		u.invoiceEditButton.Enable()
	}

	newButton := widget.NewButtonWithIcon("New Invoice", themePlusIcon(), func() {
		u.openInvoiceDialog(nil)
	})
	u.invoiceEditButton = widget.NewButton("Edit Selected", func() {
		if u.selectedInvoice < 0 || u.selectedInvoice >= len(u.invoices) {
			return
		}
		invoice := u.invoices[u.selectedInvoice]
		u.openInvoiceDialog(&invoice)
	})
	u.invoiceEditButton.Disable()

	actionBar := container.NewHBox(newButton, u.invoiceEditButton)

	split := container.NewHSplit(
		container.NewMax(u.invoiceList),
		container.NewMax(detailCard),
	)
	split.SetOffset(0.34)

	content := container.NewBorder(actionBar, nil, nil, nil, split)
	return content
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

func (u *UI) openInvoiceDialog(existing *models.Invoice) {
	isEdit := existing != nil
	title := "New Invoice"
	submitLabel := "Create"
	var current models.Invoice
	if isEdit {
		title = "Edit Invoice"
		submitLabel = "Update"
		current = *existing
	}

	profileOptions := u.profileOptions()
	customerOptions := u.customerOptions()

	profileSelect := widget.NewSelect(profileOptions, nil)
	profileSelect.PlaceHolder = "Select profile"
	customerSelect := widget.NewSelect(customerOptions, nil)
	customerSelect.PlaceHolder = "Select customer"
	issueDate := widget.NewEntry()
	issueDate.SetPlaceHolder("YYYY-MM-DD")
	dueDate := widget.NewEntry()
	dueDate.SetPlaceHolder("YYYY-MM-DD")
	taxRate := widget.NewEntry()
	taxRate.SetPlaceHolder("0")
	notes := widget.NewMultiLineEntry()
	items := make([]models.InvoiceItem, 0)

	defaultIssue := time.Now()
	defaultDue := defaultIssue.AddDate(0, 0, 14)
	if isEdit {
		if prof, err := u.store.GetProfile(current.ProfileID); err == nil {
			label := u.profileLabel(prof)
			if contains(profileOptions, label) {
				profileSelect.SetSelected(label)
			}
		}
		if cust, err := u.store.GetCustomer(current.CustomerID); err == nil {
			label := u.customerLabel(cust)
			if contains(customerOptions, label) {
				customerSelect.SetSelected(label)
			}
		}
		issueDate.SetText(current.IssueDate.Format("2006-01-02"))
		dueDate.SetText(current.DueDate.Format("2006-01-02"))
		taxRate.SetText(fmt.Sprintf("%.2f", current.TaxRatePercent))
		notes.SetText(current.Notes)
		items = append(items, current.Items...)
	} else {
		issueDate.SetText(defaultIssue.Format("2006-01-02"))
		dueDate.SetText(defaultDue.Format("2006-01-02"))
		taxRate.SetText("0")
		if len(profileOptions) > 0 {
			profileSelect.SetSelected(profileOptions[0])
		}
		if len(customerOptions) > 0 {
			customerSelect.SetSelected(customerOptions[0])
		}
		if u.lastProfileID != "" {
			if profile, ok := u.profileByID(u.lastProfileID); ok {
				label := u.profileLabel(profile)
				if contains(profileOptions, label) {
					profileSelect.SetSelected(label)
				}
			}
		}
		if u.lastCustomerID != "" {
			if customer, ok := u.customerByID(u.lastCustomerID); ok {
				label := u.customerLabel(customer)
				if contains(customerOptions, label) {
					customerSelect.SetSelected(label)
				}
			}
		}
	}

	makeHeaderLabel := func(text string) *widget.Label {
		return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}

	lineItemsContainer := container.NewVBox()
	subtotalLabel := widget.NewLabel("Subtotal: 0.00")
	taxLabel := widget.NewLabel("Tax: 0.00")
	totalLabel := widget.NewLabel("Total: 0.00")

	updateTotals := func() {
		subtotal := 0.0
		for _, item := range items {
			subtotal += item.LineTotal
		}
		taxPercent := 0.0
		if v := strings.TrimSpace(taxRate.Text); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				taxPercent = f
			}
		}
		taxAmount := subtotal * (taxPercent / 100)
		total := subtotal + taxAmount
		subtotalLabel.SetText(fmt.Sprintf("Subtotal: %.2f", subtotal))
		taxLabel.SetText(fmt.Sprintf("Tax: %.2f (%.2f%%)", taxAmount, taxPercent))
		totalLabel.SetText(fmt.Sprintf("Total: %.2f", total))
	}

	taxRate.OnChanged = func(string) {
		updateTotals()
	}

	renderLineItems := func() {}

	renderLineItems = func() {
		lineItemsContainer.Objects = nil
		if len(items) == 0 {
			lineItemsContainer.Add(widget.NewLabel("No line items yet. Use 'Add Line Item' to start."))
			updateTotals()
			lineItemsContainer.Refresh()
			return
		}
		for i := range items {
			idx := i
			item := items[idx]

			descEntry := widget.NewEntry()
			descEntry.SetText(item.Description)
			descEntry.OnChanged = func(val string) {
				items[idx].Description = val
			}

			qtyEntry := widget.NewEntry()
			qtyEntry.SetText(fmt.Sprintf("%.2f", item.Quantity))

			priceEntry := widget.NewEntry()
			priceEntry.SetText(fmt.Sprintf("%.2f", item.UnitPrice))

			totalValue := widget.NewLabel(fmt.Sprintf("%.2f", item.LineTotal))

			recalculate := func() {
				qtyVal, err := strconv.ParseFloat(strings.TrimSpace(qtyEntry.Text), 64)
				if err != nil {
					qtyVal = 0
				}
				priceVal, err := strconv.ParseFloat(strings.TrimSpace(priceEntry.Text), 64)
				if err != nil {
					priceVal = 0
				}
				items[idx].Quantity = qtyVal
				items[idx].UnitPrice = priceVal
				items[idx].LineTotal = qtyVal * priceVal
				totalValue.SetText(fmt.Sprintf("%.2f", items[idx].LineTotal))
				updateTotals()
			}
			qtyEntry.OnChanged = func(string) { recalculate() }
			priceEntry.OnChanged = func(string) { recalculate() }

			removeButton := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
				items = append(items[:idx], items[idx+1:]...)
				renderLineItems()
				updateTotals()
			})

			row := container.NewGridWithColumns(5,
				descEntry,
				qtyEntry,
				priceEntry,
				totalValue,
				removeButton,
			)
			lineItemsContainer.Add(row)
		}
		updateTotals()
		lineItemsContainer.Refresh()
	}

	addRow := func(initial models.InvoiceItem) {
		items = append(items, initial)
		renderLineItems()
	}

	addItemButton := widget.NewButtonWithIcon("Add Line Item", theme.ContentAddIcon(), func() {
		addRow(models.InvoiceItem{Description: "", Quantity: 1, UnitPrice: 0, LineTotal: 0})
	})

	headerRow := container.NewGridWithColumns(5,
		makeHeaderLabel("Description"),
		makeHeaderLabel("Qty"),
		makeHeaderLabel("Unit"),
		makeHeaderLabel("Line Total"),
		widget.NewLabel(""),
	)

	header := container.NewBorder(nil, nil, nil, addItemButton, headerRow)
	itemsScroll := container.NewVScroll(lineItemsContainer)
	itemsScroll.SetMinSize(fyne.NewSize(0, 200))

	summaryCard := widget.NewCard("Invoice Summary", "", container.NewVBox(subtotalLabel, taxLabel, totalLabel))

	if len(items) == 0 && !isEdit {
		addRow(models.InvoiceItem{Description: "", Quantity: 1, UnitPrice: 0, LineTotal: 0})
	} else {
		renderLineItems()
	}

	form := widget.NewForm(
		widget.NewFormItem("Profile", profileSelect),
		widget.NewFormItem("Customer", customerSelect),
		widget.NewFormItem("Issue Date (YYYY-MM-DD)", issueDate),
		widget.NewFormItem("Due Date (YYYY-MM-DD)", dueDate),
		widget.NewFormItem("Tax Rate (%)", taxRate),
		widget.NewFormItem("Notes", notes),
	)

	lineItemsSection := container.NewVBox(
		header,
		itemsScroll,
		widget.NewSeparator(),
		summaryCard,
	)

	content := container.NewBorder(form, nil, nil, nil, lineItemsSection)
	content = container.NewVBox(form, widget.NewSeparator(), lineItemsSection)

	var dlg *dialog.CustomDialog
	save := widget.NewButton(submitLabel, nil)
	cancel := widget.NewButton("Cancel", nil)
	status := widget.NewLabel("")
	status.Wrapping = fyne.TextWrapWord
	status.Hide()

	buttons := container.NewHBox(layout.NewSpacer(), cancel, save)
	dialogContent := container.NewBorder(content, container.NewVBox(status, buttons), nil, nil, nil)

	dlg = dialog.NewCustomWithoutButtons(title, dialogContent, u.win)

	cancel.OnTapped = func() {
		dlg.Hide()
	}

	save.OnTapped = func() {
		if len(profileOptions) == 0 || len(customerOptions) == 0 {
			status.SetText("Please create at least one profile and one customer first.")
			status.Show()
			return
		}
		profileLabel := profileSelect.Selected
		customerLabel := customerSelect.Selected
		if profileLabel == "" || customerLabel == "" {
			status.SetText("Profile and customer selection are required.")
			status.Show()
			return
		}
		profileModel, ok := u.profileByLabel(profileLabel)
		if !ok {
			status.SetText("Selected profile could not be found.")
			status.Show()
			return
		}
		customerModel, ok := u.customerByLabel(customerLabel)
		if !ok {
			status.SetText("Selected customer could not be found.")
			status.Show()
			return
		}
		if len(items) == 0 {
			status.SetText("Add at least one line item.")
			status.Show()
			return
		}
		issue, err := time.Parse("2006-01-02", strings.TrimSpace(issueDate.Text))
		if err != nil {
			status.SetText("Invalid issue date. Use YYYY-MM-DD.")
			status.Show()
			return
		}
		due, err := time.Parse("2006-01-02", strings.TrimSpace(dueDate.Text))
		if err != nil {
			status.SetText("Invalid due date. Use YYYY-MM-DD.")
			status.Show()
			return
		}
		taxPercent := 0.0
		if strings.TrimSpace(taxRate.Text) != "" {
			taxPercent, err = strconv.ParseFloat(strings.TrimSpace(taxRate.Text), 64)
			if err != nil {
				status.SetText("Tax rate must be a number.")
				status.Show()
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

		invoiceID := ""
		invoiceNumber := ""
		pdfPath := ""
		createdAt := now
		if isEdit {
			invoiceID = current.ID
			invoiceNumber = current.Number
			pdfPath = current.PDFPath
			createdAt = current.CreatedAt
		} else {
			invoiceID = id.New()
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
			ProfileID:      profileModel.ID,
			CustomerID:     customerModel.ID,
			IssueDate:      issue,
			DueDate:        due,
			Items:          append([]models.InvoiceItem(nil), items...),
			Notes:          strings.TrimSpace(notes.Text),
			TaxRatePercent: taxPercent,
			Subtotal:       subtotal,
			TaxAmount:      taxAmount,
			Total:          total,
			PDFPath:        pdfPath,
			CreatedAt:      createdAt,
			UpdatedAt:      now,
		}

		if err := u.store.SaveInvoice(invoice); err != nil {
			status.SetText(fmt.Sprintf("Save failed: %v", err))
			status.Show()
			return
		}
		if err := pdf.CreateInvoicePDF(pdfPath, profileModel, customerModel, invoice); err != nil {
			status.SetText(fmt.Sprintf("PDF generation failed: %v", err))
			status.Show()
			return
		}

		u.lastProfileID = invoice.ProfileID
		u.lastCustomerID = invoice.CustomerID
		u.refreshInvoices(invoice.ID)
		if isEdit {
			dialog.ShowInformation("Invoice updated", fmt.Sprintf("Invoice %s regenerated. PDF stored at %s.", invoice.Number, pdfPath), u.win)
		} else {
			dialog.ShowInformation("Invoice created", fmt.Sprintf("Invoice %s generated. PDF stored at %s.", invoice.Number, pdfPath), u.win)
		}
		dlg.Hide()
	}

	dlg.Resize(fyne.NewSize(560, 640))
	dlg.Show()
}

func (u *UI) showFormDialog(title, submitLabel string, form *widget.Form, submit func() error) {
	status := widget.NewLabel("")
	status.Wrapping = fyne.TextWrapWord
	status.Hide()

	var dlg *dialog.CustomDialog
	save := widget.NewButton(submitLabel, nil)
	cancel := widget.NewButton("Cancel", nil)
	buttons := container.NewHBox(layout.NewSpacer(), cancel, save)
	content := container.NewVBox(form, status, buttons)

	dlg = dialog.NewCustomWithoutButtons(title, content, u.win)

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

func (u *UI) refreshProfiles(selectedIDs ...string) {
	profiles, err := u.store.ListProfiles()
	if err != nil {
		dialog.ShowError(fmt.Errorf("load profiles: %w", err), u.win)
		return
	}
	targetID := ""
	if len(selectedIDs) > 0 {
		targetID = selectedIDs[0]
	} else if u.selectedProfile >= 0 && u.selectedProfile < len(u.profiles) {
		targetID = u.profiles[u.selectedProfile].ID
	}

	u.profiles = profiles
	u.selectedProfile = -1

	if u.profileList != nil {
		u.profileList.Refresh()
	}

	if targetID != "" {
		for idx, profile := range profiles {
			if profile.ID == targetID {
				u.selectedProfile = idx
				break
			}
		}
	}

	if u.profileEditButton != nil {
		if u.selectedProfile >= 0 {
			u.profileEditButton.Enable()
		} else {
			u.profileEditButton.Disable()
		}
	}

	if u.selectedProfile >= 0 && u.profileList != nil && u.selectedProfile < len(u.profiles) {
		u.profileList.Select(u.selectedProfile)
	} else {
		u.updateProfileDetail()
	}
}

func (u *UI) refreshCustomers(selectedIDs ...string) {
	customers, err := u.store.ListCustomers()
	if err != nil {
		dialog.ShowError(fmt.Errorf("load customers: %w", err), u.win)
		return
	}
	targetID := ""
	if len(selectedIDs) > 0 {
		targetID = selectedIDs[0]
	} else if u.selectedCustomer >= 0 && u.selectedCustomer < len(u.customers) {
		targetID = u.customers[u.selectedCustomer].ID
	}

	u.customers = customers
	u.selectedCustomer = -1

	if u.customerList != nil {
		u.customerList.Refresh()
	}

	if targetID != "" {
		for idx, customer := range customers {
			if customer.ID == targetID {
				u.selectedCustomer = idx
				break
			}
		}
	}

	if u.customerEditButton != nil {
		if u.selectedCustomer >= 0 {
			u.customerEditButton.Enable()
		} else {
			u.customerEditButton.Disable()
		}
	}

	if u.selectedCustomer >= 0 && u.customerList != nil && u.selectedCustomer < len(u.customers) {
		u.customerList.Select(u.selectedCustomer)
	} else {
		u.updateCustomerDetail()
	}
}

func (u *UI) refreshInvoices(selectedIDs ...string) {
	invoices, err := u.store.ListInvoices()
	if err != nil {
		dialog.ShowError(fmt.Errorf("load invoices: %w", err), u.win)
		return
	}
	targetID := ""
	if len(selectedIDs) > 0 {
		targetID = selectedIDs[0]
	} else if u.selectedInvoice >= 0 && u.selectedInvoice < len(u.invoices) {
		targetID = u.invoices[u.selectedInvoice].ID
	}

	u.invoices = invoices
	u.invoiceSummaries = make([]string, len(invoices))
	for idx, inv := range invoices {
		customerLabel := inv.CustomerID
		if cust, err := u.store.GetCustomer(inv.CustomerID); err == nil {
			customerLabel = cust.DisplayName
		}
		status := invoiceBadge(inv)
		u.invoiceSummaries[idx] = fmt.Sprintf("%s | %s – %s – Total %.2f", status, inv.Number, customerLabel, inv.Total)
	}

	u.selectedInvoice = -1

	if u.invoiceList != nil {
		u.invoiceList.Refresh()
	}

	if targetID != "" {
		for idx, invoice := range invoices {
			if invoice.ID == targetID {
				u.selectedInvoice = idx
				break
			}
		}
	}

	if u.invoiceEditButton != nil {
		if u.selectedInvoice >= 0 {
			u.invoiceEditButton.Enable()
		} else {
			u.invoiceEditButton.Disable()
		}
	}

	if u.selectedInvoice >= 0 && u.invoiceList != nil && u.selectedInvoice < len(u.invoices) {
		u.invoiceList.Select(u.selectedInvoice)
	} else {
		u.updateInvoiceDetail()
	}
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

func (u *UI) updateInvoiceDetail() {
	if u.invoiceDetailText == nil {
		return
	}
	if u.selectedInvoice < 0 || u.selectedInvoice >= len(u.invoices) {
		u.invoiceDetailText.ParseMarkdown("_Select an invoice to view details._")
		return
	}
	inv := u.invoices[u.selectedInvoice]
	customerName := inv.CustomerID
	if cust, err := u.store.GetCustomer(inv.CustomerID); err == nil {
		customerName = cust.DisplayName
	}
	profileName := inv.ProfileID
	if prof, err := u.store.GetProfile(inv.ProfileID); err == nil {
		profileName = prof.DisplayName
	}
	status := invoiceBadge(inv)
	now := time.Now()
	daysUntilDue := int(inv.DueDate.Sub(now).Hours() / 24)
	dueDescriptor := ""
	if daysUntilDue < 0 {
		dueDescriptor = fmt.Sprintf("Overdue by %d days", -daysUntilDue)
	} else {
		dueDescriptor = fmt.Sprintf("Due in %d days", daysUntilDue)
	}

	lines := []string{
		fmt.Sprintf("**Invoice:** %s", inv.Number),
		fmt.Sprintf("**Status:** %s (%s)", status, dueDescriptor),
		fmt.Sprintf("**Profile:** %s", profileName),
		fmt.Sprintf("**Customer:** %s", customerName),
		fmt.Sprintf("**Issued:** %s", inv.IssueDate.Format("2006-01-02")),
		fmt.Sprintf("**Due:** %s", inv.DueDate.Format("2006-01-02")),
		fmt.Sprintf("**Subtotal:** %.2f", inv.Subtotal),
		fmt.Sprintf("**Tax:** %.2f (%.2f%%)", inv.TaxAmount, inv.TaxRatePercent),
		fmt.Sprintf("**Total:** %.2f", inv.Total),
		fmt.Sprintf("**PDF:** %s", inv.PDFPath),
		"",
		"**Line Items**",
	}
	for _, item := range inv.Items {
		lines = append(lines, fmt.Sprintf("- %s: %.2f × %.2f = %.2f", item.Description, item.Quantity, item.UnitPrice, item.LineTotal))
	}
	if strings.TrimSpace(inv.Notes) != "" {
		lines = append(lines, "", "**Notes**", inv.Notes)
	}
	u.invoiceDetailText.ParseMarkdown(strings.Join(lines, "\n"))
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
	now := time.Now()
	if inv.DueDate.Before(now) {
		return "⛔ Overdue"
	}
	if inv.DueDate.Sub(now) <= 72*time.Hour {
		return "⚠️ Due soon"
	}
	return "✅ On track"
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func themePlusIcon() fyne.Resource {
	return theme.ContentAddIcon()
}
