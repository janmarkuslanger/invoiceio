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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/janmarkuslanger/invoiceio/internal/id"
	"github.com/janmarkuslanger/invoiceio/internal/models"
	"github.com/janmarkuslanger/invoiceio/internal/pdf"
)

func (u *UI) makeInvoicesTab() fyne.CanvasObject {
	u.invoiceDetailText = widget.NewRichTextFromMarkdown("_Select an invoice to view details._")
	u.invoiceDetailText.Wrapping = fyne.TextWrapWord
	detailCard := widget.NewCard("Invoice Details", "", u.invoiceDetailText)
	detailScroll := container.NewVScroll(detailCard)

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
		container.NewMax(detailScroll),
	)
	split.SetOffset(0.34)

	return container.NewBorder(actionBar, nil, nil, nil, split)
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

	var renderLineItems func()
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

	content := container.NewVBox(form, widget.NewSeparator(), lineItemsSection)

	save := widget.NewButton(submitLabel, nil)
	cancel := widget.NewButton("Cancel", nil)
	status := widget.NewLabel("")
	status.Wrapping = fyne.TextWrapWord
	status.Hide()

	buttons := container.NewHBox(layout.NewSpacer(), cancel, save)
	dialogContent := container.NewBorder(content, container.NewVBox(status, buttons), nil, nil, nil)

	dlg := dialog.NewCustomWithoutButtons(title, dialogContent, u.win)

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
