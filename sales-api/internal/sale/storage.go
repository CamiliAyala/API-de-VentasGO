package sale

import "errors"

// ErrNotFound is returned when a user with the given ID is not found.
var ErrNotFound = errors.New("user not found")
var ErrNotFoundSale = errors.New("sale not found")

// ErrEmptyID is returned when trying to store a user with an empty ID.
var ErrEmptyID = errors.New("empty user ID")

// Storage is the main interface for our storage layer.
type Storage interface {
	SetSale(sale *Sale) error
	ReadSale(id string) (*Sale, error)
	ReadAllSales() (map[string]*Sale, error)
}

// LocalStorage provides an in-memory implementation for storing users.
type LocalStorage struct {
	s map[string]*Sale
}

// NewLocalStorage instantiates a new LocalStorage with an empty map.
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		s: map[string]*Sale{},
	}
}

// Set stores or updates a user in the local storage.

func (l *LocalStorage) SetSale(sale *Sale) error {
	if sale.ID == "" {
		return ErrEmptyID
	}

	l.s[sale.ID] = sale
	return nil
}

// Read retrieves a sale from the local storage by ID.

func (l *LocalStorage) ReadSale(id string) (*Sale, error) {
	u, ok := l.s[id]
	if !ok {
		return nil, ErrNotFoundSale
	}
	return u, nil
}

func (l *LocalStorage) ReadAllSales() (map[string]*Sale, error) {
	u := l.s
	if len(u) == 0 {
		return nil, ErrNotFoundSale
	}
	return u, nil
}
