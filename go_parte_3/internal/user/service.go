package user

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	letterRegex = regexp.MustCompile(`^[A-Za-z]+$`)

	ErrInvalidInput     = errors.New("invalid input")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
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

// Create adds a brand-new user to the system.
// It sets CreatedAt and UpdatedAt to the current time and initializes Version to 1.
// Returns ErrEmptyID if user.ID is empty.
func (s *Service) Create(user *User) error {

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

	if err := s.storage.Set(user); err != nil {
		s.logger.Error("failed to set user", zap.Error(err), zap.Any("user", user))
		return err
	}

	return nil
}

// Get retrieves a user by its ID.
// Returns ErrNotFound if no user exists with the given ID.
/*func (s *Service) Get(id string) (*User, error) {
	return s.storage.Read(id)
}*/

// si el users esta inactivo, no lo devolverá
func (s *Service) Get(id string) (*User, error) {
	user, err := s.storage.Read(id)
	if err != nil {
		return nil, err
	}
	/*if user.Status == UserStatusDeleted {
		return nil, ErrNotFound
	}*/
	return user, nil
}

// Update modifies an existing user's data.
// It updates Name, Address, NickName, sets UpdatedAt to now and increments Version.
// Returns ErrNotFound if the user does not exist, or ErrEmptyID if user.ID is empty.
/*func (s *Service) Update(id string, user *UpdateFields) (*User, error) {
	existing, err := s.storage.Read(id)
	if err != nil {
		return nil, err
	}

	if user.Name != nil {
		existing.Name = *user.Name
	}

	if user.Address != nil {
		existing.Address = *user.Address
	}

	if user.NickName != nil {
		existing.NickName = *user.NickName
	}

	existing.UpdatedAt = time.Now()
	existing.Version++

	if err := s.storage.Set(existing); err != nil {
		return nil, err
	}

	return existing, nil
}*/

//Update
//Si no se modifica ningún valor debe arrojar un 400.
//Los campos deben respetar lo mismo del create.

func (s *Service) Update(id string, updates *UpdateFields) (*User, error) {
	existing, err := s.storage.Read(id)
	if err != nil {
		return nil, err
	}

	// Chequear si hay cambios
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

	if err := s.storage.Set(existing); err != nil {
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
	user, err := s.storage.Read(id)
	if err != nil {
		return err
	}

	user.Status = UserStatusDeleted
	user.UpdatedAt = time.Now()
	user.Version++

	if err := s.storage.Set(user); err != nil {
		s.logger.Error("failed to set user as deleted", zap.Error(err), zap.String("id", id))
		return err
	}

	return nil
}

const (
	UserStatusActive  = "active"
	UserStatusDeleted = "deleted"
)
