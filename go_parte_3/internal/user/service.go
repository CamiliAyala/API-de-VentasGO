package user

import (
	"errors"
	"math/rand"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	letterRegex           = regexp.MustCompile(`^[A-Za-z]+$`)
	ErrInvalidInput       = errors.New("invalid input")
	ErrNoFieldsToUpdate   = errors.New("no fields to update")
	ErrUserNotFound       = errors.New("user not found")
	ErrSaleNotFound       = errors.New("sale not found")
	ErrTransactionInvalid = errors.New("transaccion invalida")
)

// Service provides high-level user management operations on a LocalStorage backend.
type Service struct {
	// storage is the underlying persistence for User entities.
	storage Storage

	// logger is our observability component to log.
	logger *zap.Logger
}

// NewService creates a new Service.
func NewService(storage Storage, logger *zap.Logger) *Service {
	if logger == nil {
		logger, _ = zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any
	}

	return &Service{
		storage: storage,
		logger:  logger,
	}
}

// Create de Usuario
func (s *Service) CreateUser(user *User) error {

	if user.Name == "" || user.Address == "" {
		return ErrInvalidInput
	}
	if !letterRegex.MatchString(user.Name) {
		return ErrInvalidInput
	}
	if user.NickName != "" && !letterRegex.MatchString(user.NickName) {
		return ErrInvalidInput
	}

	user.ID = uuid.NewString()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Version = 1
	user.Status = UserStatusActive

	if err := s.storage.SetUser(user); err != nil {
		s.logger.Error("failed to set user", zap.Error(err), zap.Any("user", user))
		return err
	}

	return nil
}

func (s *Service) CreateSale(sale *Sale) error {
	u, _ := s.GetUser(sale.UserID)
	if u == nil {
		return ErrSaleNotFound
	}
	if sale.Amount == 0.0 {
		return ErrInvalidInput
	}

	sale.ID = uuid.NewString()
	sale.UserID = u.ID
	statuses := []string{"pending", "approved", "rejected"}
	sale.Status = statuses[rand.Intn(len(statuses))]
	now := time.Now()
	sale.CreatedAt = now
	sale.UpdatedAt = now
	sale.Version = 1

	if err := s.storage.SetSale(sale); err != nil {
		s.logger.Error("failed to set sale", zap.Error(err), zap.Any("sale", sale))
		return err
	}

	return nil
}

// si el users esta inactivo, no lo devolverá
func (s *Service) GetUser(id string) (*User, error) {
	user, err := s.storage.ReadUser(id)
	if err != nil {
		return nil, err
	}
	if user.Status == UserStatusDeleted {
		return nil, ErrNotFound
	}
	return user, nil
}
func (s *Service) GetSale(id string) (*Sale, error) {
	sale, err := s.storage.ReadSale(id)
	if err != nil {
		return nil, err
	}
	return sale, nil
}

func (s *Service) GetSaleByUserAndStatus(userID string, status string) (informe, error) {
	var resp informe
	var meta metadata
	resp.Results = []Sale{}

	// Validar estado si se envía
	if status != "" && status != "approved" && status != "rejected" && status != "pending" {
		return resp, ErrInvalidInput
	}

	salesMap, err := s.storage.ReadAllSales()
	if err != nil {
		return resp, err
	}

	var filteredSales []Sale
	for _, sale := range salesMap {
		if sale.UserID != userID {
			continue
		}
		if status == "" || sale.Status == status {
			filteredSales = append(filteredSales, *sale)
		}
	}

	meta.Quantity = len(filteredSales)
	for _, sale := range filteredSales {
		switch sale.Status {
		case "approved":
			meta.Approved++
		case "rejected":
			meta.Rejected++
		case "pending":
			meta.Pending++
		}
		meta.TotalAmount += sale.Amount
	}

	resp.Metadata = meta
	resp.Results = filteredSales
	return resp, nil
}

//Update
//Si no se modifica ningún valor debe arrojar un 400.

func (s *Service) UpdateUser(id string, updates *UpdateFieldsUser) (*User, error) {
	existing, err := s.storage.ReadUser(id)
	if err != nil {
		return nil, err
	}

	updated := false

	if updates.Name != nil {
		if !letterRegex.MatchString(*updates.Name) {
			return nil, ErrInvalidInput
		}
		existing.Name = *updates.Name
		updated = true
	}

	if updates.Address != nil {
		// No validamos Address por letras, solo permitimos actualizar
		existing.Address = *updates.Address
		updated = true
	}

	if updates.NickName != nil {
		if !letterRegex.MatchString(*updates.NickName) {
			return nil, ErrInvalidInput
		}
		existing.NickName = *updates.NickName
		updated = true
	}

	// Si no se modificó nada, lanzar error 400
	if !updated {
		return nil, ErrNoFieldsToUpdate
	}

	existing.UpdatedAt = time.Now()
	existing.Version++

	if err := s.storage.SetUser(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *Service) UpdateSale(id string, updates *UpdateFieldsSale) (*Sale, error) {
	existing, err := s.storage.ReadSale(id)

	if err != nil {
		return nil, err
	}

	// Chequear si hay cambios
	updated := false

	if existing.Status == "pending" {
		if updates.Status == "rejected" || updates.Status == "approved" {
			existing.Status = updates.Status
			updated = true
		} else {
			return nil, ErrInvalidInput
		}
	} else {
		return nil, ErrTransactionInvalid
	}

	// Si no se modificó nada, lanzar error 400
	if !updated {
		return nil, ErrNoFieldsToUpdate
	}

	existing.UpdatedAt = time.Now()
	existing.Version++

	if err := s.storage.SetSale(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes a user from the system by its ID.
// Returns ErrNotFound if the user does not exist.
/*func (s *Service) Delete(id string) error {
	return s.storage.Delete(id)
}*/

// Hacer que el borrado sea lógico en vez de físico.
func (s *Service) Delete(id string) error {
	user, err := s.storage.ReadUser(id)
	if err != nil {
		return err
	}

	user.Status = UserStatusDeleted
	user.UpdatedAt = time.Now()
	user.Version++

	if err := s.storage.SetUser(user); err != nil {
		s.logger.Error("failed to set user as deleted", zap.Error(err), zap.String("id", id))
		return err
	}

	return nil
}

const (
	UserStatusActive  = "active"
	UserStatusDeleted = "deleted"
)
