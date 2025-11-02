package ui

import "fmt"

func (u *UI) refreshProfiles(selectedIDs ...string) {
	profiles, err := u.store.ListProfiles()
	if err != nil {
		dialogError(u.win, fmt.Errorf("load profiles: %w", err))
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
		dialogError(u.win, fmt.Errorf("load customers: %w", err))
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
		dialogError(u.win, fmt.Errorf("load invoices: %w", err))
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
