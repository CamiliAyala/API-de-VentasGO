package user

import (
	"errors"
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
