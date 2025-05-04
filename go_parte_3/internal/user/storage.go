package user

import "errors"

// ErrNotFound is returned when a user with the given ID is not found.
var ErrNotFound = errors.New("user not found")
var ErrNotFoundSale = errors.New("sale not found")

// ErrEmptyID is returned when trying to store a user with an empty ID.
var ErrEmptyID = errors.New("empty user ID")

// Storage is the main interface for our storage layer.
type Storage interface {
	SetUser(user *User) error
	SetSale(sale *Sale) error
	ReadUser(id string) (*User, error)
	ReadSale(id string) (*Sale, error)
	ReadAllSales() (map[string]*Sale, error)
	Delete(id string) error
}

// LocalStorage provides an in-memory implementation for storing users.
type LocalStorage struct {
	m map[string]*User
	s map[string]*Sale
}

// NewLocalStorage instantiates a new LocalStorage with an empty map.
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		m: map[string]*User{},
		s: map[string]*Sale{},
	}
}

// Set stores or updates a user in the local storage.
// Returns ErrEmptyID if the user has an empty ID.
func (l *LocalStorage) SetUser(user *User) error {
	if user.ID == "" {
		return ErrEmptyID
	}

	l.m[user.ID] = user
	return nil
}

func (l *LocalStorage) SetSale(sale *Sale) error {
	if sale.ID == "" {
		return ErrEmptyID
	}

	l.s[sale.ID] = sale
	return nil
}

// Read retrieves a user from the local storage by ID.
// Returns ErrNotFound if the user is not found.
func (l *LocalStorage) ReadUser(id string) (*User, error) {
	u, ok := l.m[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

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

// Delete removes a user from the local storage by ID.
// Returns ErrNotFound if the user does not exist.
func (l *LocalStorage) Delete(id string) error {
	_, err := l.ReadUser(id)
	if err != nil {
		return err
	}

	delete(l.m, id)
	return nil
}
