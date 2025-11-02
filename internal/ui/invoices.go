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

	"github.com/janmarkuslanger/invoiceio/internal/i18n"
	"github.com/janmarkuslanger/invoiceio/internal/id"
	"github.com/janmarkuslanger/invoiceio/internal/models"
	"github.com/janmarkuslanger/invoiceio/internal/pdf"
)

func (u *UI) makeInvoicesTab() fyne.CanvasObject {
	u.invoiceDetailText = widget.NewRichTextFromMarkdown(i18n.T("invoices.detail.empty"))
	u.invoiceDetailText.Wrapping = fyne.TextWrapWord
	detailCard := widget.NewCard(i18n.T("invoices.detail.title"), "", u.invoiceDetailText)
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

	newButton := widget.NewButtonWithIcon(i18n.T("invoices.button.new"), themePlusIcon(), func() {
		u.openInvoiceDialog(nil)
	})
	u.invoiceEditButton = widget.NewButton(i18n.T("button.editSelected"), func() {
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
	title := i18n.T("invoices.dialog.newTitle")
	submitLabel := i18n.T("invoices.dialog.create")
	var current models.Invoice
	if isEdit {
		title = i18n.T("invoices.dialog.editTitle")
		submitLabel = i18n.T("invoices.dialog.update")
		current = *existing
	}

	profileOptions := u.profileOptions()
	customerOptions := u.customerOptions()

	profileSelect := widget.NewSelect(profileOptions, nil)
	profileSelect.PlaceHolder = i18n.T("invoices.form.profilePlaceholder")
	customerSelect := widget.NewSelect(customerOptions, nil)
	customerSelect.PlaceHolder = i18n.T("invoices.form.customerPlaceholder")
	issueDate := widget.NewEntry()
	issueDate.SetPlaceHolder(i18n.T("invoices.form.issueDatePlaceholder"))
	dueDate := widget.NewEntry()
	dueDate.SetPlaceHolder(i18n.T("invoices.form.dueDatePlaceholder"))
	taxRate := widget.NewEntry()
	taxRate.SetPlaceHolder(i18n.T("invoices.form.taxRatePlaceholder"))
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
	subtotalLabel := widget.NewLabel("")
	taxLabel := widget.NewLabel("")
	totalLabel := widget.NewLabel("")

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
		subtotalLabel.SetText(i18n.T("invoices.summary.subtotal", subtotal))
		taxLabel.SetText(i18n.T("invoices.summary.tax", taxAmount, taxPercent))
		totalLabel.SetText(i18n.T("invoices.summary.total", total))
	}

	taxRate.OnChanged = func(string) {
		updateTotals()
	}

	var renderLineItems func()
	renderLineItems = func() {
		lineItemsContainer.Objects = nil
		if len(items) == 0 {
			lineItemsContainer.Add(widget.NewLabel(i18n.T("invoices.lineItems.empty", i18n.T("invoices.lineItems.add"))))
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

	addItemButton := widget.NewButtonWithIcon(i18n.T("invoices.lineItems.add"), theme.ContentAddIcon(), func() {
		addRow(models.InvoiceItem{Description: "", Quantity: 1, UnitPrice: 0, LineTotal: 0})
	})

	headerRow := container.NewGridWithColumns(5,
		makeHeaderLabel(i18n.T("invoices.table.description")),
		makeHeaderLabel(i18n.T("invoices.table.quantity")),
		makeHeaderLabel(i18n.T("invoices.table.unit")),
		makeHeaderLabel(i18n.T("invoices.table.lineTotal")),
		widget.NewLabel(""),
	)

	header := container.NewBorder(nil, nil, nil, addItemButton, headerRow)
	itemsScroll := container.NewVScroll(lineItemsContainer)
	itemsScroll.SetMinSize(fyne.NewSize(0, 200))

	summaryCard := widget.NewCard(i18n.T("invoices.summary.title"), "", container.NewVBox(subtotalLabel, taxLabel, totalLabel))

	if len(items) == 0 && !isEdit {
		addRow(models.InvoiceItem{Description: "", Quantity: 1, UnitPrice: 0, LineTotal: 0})
	} else {
		renderLineItems()
	}

	form := widget.NewForm(
		widget.NewFormItem(i18n.T("invoices.form.profile"), profileSelect),
		widget.NewFormItem(i18n.T("invoices.form.customer"), customerSelect),
		widget.NewFormItem(i18n.T("invoices.form.issueDate"), issueDate),
		widget.NewFormItem(i18n.T("invoices.form.dueDate"), dueDate),
		widget.NewFormItem(i18n.T("invoices.form.taxRate"), taxRate),
		widget.NewFormItem(i18n.T("invoices.form.notes"), notes),
	)

	lineItemsSection := container.NewVBox(
		header,
		itemsScroll,
		widget.NewSeparator(),
		summaryCard,
	)

	content := container.NewVBox(form, widget.NewSeparator(), lineItemsSection)

	save := widget.NewButton(submitLabel, nil)
	cancel := widget.NewButton(i18n.T("common.cancel"), nil)
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
			status.SetText(i18n.T("invoices.error.setupRequired"))
			status.Show()
			return
		}
		profileLabel := profileSelect.Selected
		customerLabel := customerSelect.Selected
		if profileLabel == "" || customerLabel == "" {
			status.SetText(i18n.T("invoices.error.selectionRequired"))
			status.Show()
			return
		}
		profileModel, ok := u.profileByLabel(profileLabel)
		if !ok {
			status.SetText(i18n.T("invoices.error.profileMissing"))
			status.Show()
			return
		}
		customerModel, ok := u.customerByLabel(customerLabel)
		if !ok {
			status.SetText(i18n.T("invoices.error.customerMissing"))
			status.Show()
			return
		}
		if len(items) == 0 {
			status.SetText(i18n.T("invoices.error.noItems"))
			status.Show()
			return
		}
		issue, err := time.Parse("2006-01-02", strings.TrimSpace(issueDate.Text))
		if err != nil {
			status.SetText(i18n.T("invoices.error.issueDate"))
			status.Show()
			return
		}
		due, err := time.Parse("2006-01-02", strings.TrimSpace(dueDate.Text))
		if err != nil {
			status.SetText(i18n.T("invoices.error.dueDate"))
			status.Show()
			return
		}
		taxPercent := 0.0
		if strings.TrimSpace(taxRate.Text) != "" {
			taxPercent, err = strconv.ParseFloat(strings.TrimSpace(taxRate.Text), 64)
			if err != nil {
				status.SetText(i18n.T("invoices.error.taxRate"))
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
			status.SetText(i18n.T("invoices.error.saveFailed", err))
			status.Show()
			return
		}
		if err := pdf.CreateInvoicePDF(pdfPath, profileModel, customerModel, invoice); err != nil {
			status.SetText(i18n.T("invoices.error.pdfFailed", err))
			status.Show()
			return
		}

		u.lastProfileID = invoice.ProfileID
		u.lastCustomerID = invoice.CustomerID
		u.refreshInvoices(invoice.ID)
		if isEdit {
			dialog.ShowInformation(i18n.T("invoices.info.updatedTitle"), i18n.T("invoices.info.updatedBody", invoice.Number, pdfPath), u.win)
		} else {
			dialog.ShowInformation(i18n.T("invoices.info.createdTitle"), i18n.T("invoices.info.createdBody", invoice.Number, pdfPath), u.win)
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
		u.invoiceDetailText.ParseMarkdown(i18n.T("invoices.detail.empty"))
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
		dueDescriptor = i18n.T("invoices.due.overdueBy", -daysUntilDue)
	} else {
		dueDescriptor = i18n.T("invoices.due.inDays", daysUntilDue)
	}

	lines := []string{
		i18n.T("invoices.detail.invoice", inv.Number),
		i18n.T("invoices.detail.status", status, dueDescriptor),
		i18n.T("invoices.detail.profile", profileName),
		i18n.T("invoices.detail.customer", customerName),
		i18n.T("invoices.detail.issued", inv.IssueDate.Format("2006-01-02")),
		i18n.T("invoices.detail.due", inv.DueDate.Format("2006-01-02")),
		i18n.T("invoices.detail.subtotal", inv.Subtotal),
		i18n.T("invoices.detail.tax", inv.TaxAmount, inv.TaxRatePercent),
		i18n.T("invoices.detail.total", inv.Total),
		i18n.T("invoices.detail.pdf", inv.PDFPath),
		"",
		i18n.T("invoices.detail.lineItems"),
	}
	for _, item := range inv.Items {
		lines = append(lines, i18n.T("invoices.detail.lineItem", item.Description, item.Quantity, item.UnitPrice, item.LineTotal))
	}
	if strings.TrimSpace(inv.Notes) != "" {
		lines = append(lines, "", i18n.T("invoices.detail.notesTitle"), inv.Notes)
	}
	u.invoiceDetailText.ParseMarkdown(strings.Join(lines, "\n"))
}
