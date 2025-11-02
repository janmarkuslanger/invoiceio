package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/janmarkuslanger/jsonstore"

	"github.com/janmarkuslanger/invoiceio/internal/models"
)

// Storage wires together the type-safe json stores that keep the application data.
type Storage struct {
	baseDir       string
	profileStore  *jsonstore.Store[models.Profile]
	customerStore *jsonstore.Store[models.Customer]
	invoiceStore  *jsonstore.Store[models.Invoice]
}

// ErrNotFound is returned when an entity can not be located in the underlying store.
var ErrNotFound = jsonstore.ErrNotFound

// New initialises the storage layer inside the provided base directory, creating it if required.
func New(baseDir string) (*Storage, error) {
	if baseDir == "" {
		return nil, errors.New("storage: base directory is required")
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: create base directory: %w", err)
	}

	profiles, err := jsonstore.NewStore[models.Profile](filepath.Join(baseDir, "profiles.json"))
	if err != nil {
		return nil, fmt.Errorf("storage: open profiles store: %w", err)
	}
	customers, err := jsonstore.NewStore[models.Customer](filepath.Join(baseDir, "customers.json"))
	if err != nil {
		return nil, fmt.Errorf("storage: open customers store: %w", err)
	}
	invoices, err := jsonstore.NewStore[models.Invoice](filepath.Join(baseDir, "invoices.json"))
	if err != nil {
		return nil, fmt.Errorf("storage: open invoices store: %w", err)
	}

	return &Storage{
		baseDir:       baseDir,
		profileStore:  profiles,
		customerStore: customers,
		invoiceStore:  invoices,
	}, nil
}

func (s *Storage) SaveProfile(p models.Profile) error {
	p.UpdatedAt = time.Now()
	return s.profileStore.Set(p.ID, p)
}

func (s *Storage) GetProfile(id string) (models.Profile, error) {
	return s.profileStore.Get(id)
}

func (s *Storage) DeleteProfile(id string) error {
	return s.profileStore.Delete(id)
}

func (s *Storage) ListProfiles() ([]models.Profile, error) {
	return listAll(s.profileStore, func(p models.Profile) string { return p.DisplayName })
}

func (s *Storage) SaveCustomer(c models.Customer) error {
	c.UpdatedAt = time.Now()
	return s.customerStore.Set(c.ID, c)
}

func (s *Storage) GetCustomer(id string) (models.Customer, error) {
	return s.customerStore.Get(id)
}

func (s *Storage) DeleteCustomer(id string) error {
	return s.customerStore.Delete(id)
}

func (s *Storage) ListCustomers() ([]models.Customer, error) {
	return listAll(s.customerStore, func(c models.Customer) string { return c.DisplayName })
}

func (s *Storage) SaveInvoice(inv models.Invoice) error {
	inv.UpdatedAt = time.Now()
	return s.invoiceStore.Set(inv.ID, inv)
}

func (s *Storage) GetInvoice(id string) (models.Invoice, error) {
	return s.invoiceStore.Get(id)
}

func (s *Storage) DeleteInvoice(id string) error {
	return s.invoiceStore.Delete(id)
}

func (s *Storage) ListInvoices() ([]models.Invoice, error) {
	return listAll(s.invoiceStore, func(inv models.Invoice) string {
		return fmt.Sprintf("%s-%s", inv.IssueDate.Format(time.RFC3339), inv.Number)
	})
}

// BaseDir returns the root directory that contains the json files.
func (s *Storage) BaseDir() string {
	return s.baseDir
}

func listAll[T any](store *jsonstore.Store[T], sortKey func(T) string) ([]T, error) {
	keys, err := store.Keys()
	if err != nil {
		return nil, err
	}
	sort.Strings(keys)

	out := make([]T, 0, len(keys))
	for _, key := range keys {
		v, err := store.Get(key)
		if err != nil {
			if errors.Is(err, jsonstore.ErrNotFound) {
				continue
			}
			return nil, err
		}
		out = append(out, v)
	}

	sort.SliceStable(out, func(i, j int) bool {
		return sortKey(out[i]) < sortKey(out[j])
	})
	return out, nil
}
