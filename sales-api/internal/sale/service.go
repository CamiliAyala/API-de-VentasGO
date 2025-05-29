package sale

import (
	"errors"
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrNoFieldsToUpdate   = errors.New("no fields to update")
	ErrUserNotFound       = errors.New("user not found")
	ErrSaleNotFound       = errors.New("sale not found")
	ErrTransactionInvalid = errors.New("transicion invalida")
	ErrNotUserFound       = errors.New("not user found")
	ErrTryingToGetUser    = errors.New("error trying to get user")
)

// Service provides high-level user management operations on a LocalStorage backend.
type Service struct {
	storage    Storage
	logger     *zap.Logger
	userClient *resty.Client
	urlUser    string
}

// NewService creates a new Service.
func NewService(storage Storage, logger *zap.Logger, urlUser string) *Service {
	if logger == nil {
		logger, _ = zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any
	}
	restyClient := resty.New()

	return &Service{
		storage:    storage,
		logger:     logger,
		userClient: restyClient,
		urlUser:    urlUser,
	}
}

// CreateSale creates a new sale in the system.
func (s *Service) CreateSale(sale *Sale) error {
	if sale.Amount <= 0.0 {
		return ErrInvalidInput
	}
	sale.ID = uuid.NewString()
	statuses := []string{"pending", "rejected"}
	sale.Status = statuses[rand.Intn(len(statuses))]
	now := time.Now()
	sale.CreatedAt = now
	sale.UpdatedAt = now
	sale.Version = 1
	userID := sale.UserID
	res, err := s.userClient.R().Get(s.urlUser + "/users/" + userID)
	if err != nil {
		return ErrTryingToGetUser
	}

	if res.IsError() {
		return ErrUserNotFound
	}

	if err := s.storage.SetSale(sale); err != nil {
		s.logger.Error("failed to set sale", zap.Error(err), zap.Any("sale", sale))
		return err
	}

	return nil
}

// GetUser retrieves a user by its ID.
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

	salesMap, _ := s.storage.ReadAllSales()

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
	if len(filteredSales) == 0 {
		resp.Results = []Sale{}
	} else {
		resp.Results = filteredSales
	}
	return resp, nil
}

//UpdateSale updates a sale in the system.

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
		if updates.Status == "rejected" || updates.Status == "approved" {
			return nil, ErrTransactionInvalid
		} else {
			return nil, ErrInvalidInput
		}

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
