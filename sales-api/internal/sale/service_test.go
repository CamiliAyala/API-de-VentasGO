package sale

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestService_Create_Simple(t *testing.T) {
	// Mock externo para la API de usuarios
	mockHandler := http.NewServeMux()
	mockHandler.HandleFunc("/users/1234", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "1234"}`))
	})
	mockServer := httptest.NewServer(mockHandler)
	defer mockServer.Close()

	// Creamos el servicio usando el mock como baseURL
	s := NewService(NewLocalStorage(), nil, mockServer.URL)

	input := &Sale{
		UserID: "1234", // simulamos que el UserID fue validado
		Amount: 100.0,
	}

	err := s.CreateSale(input)

	require.Nil(t, err)
	require.Equal(t, "1234", input.UserID)
	require.NotEmpty(t, input.ID)
	require.NotEmpty(t, input.Amount)
	require.NotEmpty(t, input.CreatedAt)
	require.NotEmpty(t, input.UpdatedAt)
	require.Equal(t, 1, input.Version)

	// Caso de error al guardar
	s = NewService(&mockStorage{
		mockSetSale: func(sale *Sale) error {
			return errors.New("fake error trying to set sale")
		},
	}, nil, mockServer.URL)

	err = s.CreateSale(input)
	require.NotNil(t, err)
	require.EqualError(t, err, "fake error trying to set sale")
}

func TestService_Create(t *testing.T) {
	type fields struct {
		storage Storage
		apiURL  string
	}

	type args struct {
		sale *Sale
	}

	mockHandler := http.NewServeMux()

	// Usuario válido
	mockHandler.HandleFunc("/users/1234", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"1234"}`))
	})

	// Usuario inexistente
	mockHandler.HandleFunc("/users/9999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	mockServer := httptest.NewServer(mockHandler)
	defer mockServer.Close()

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  func(t *testing.T, err error)
		wantSale func(t *testing.T, sale *Sale)
	}{
		{
			name: "error set sale",
			fields: fields{
				storage: &mockStorage{
					mockSetSale: func(sale *Sale) error {
						return errors.New("fake error trying to set sale")
					},
				},
				apiURL: mockServer.URL,
			},
			args: args{
				sale: &Sale{UserID: "1234", Amount: 100.0},
			},
			wantErr: func(t *testing.T, err error) {
				require.NotNil(t, err)
				require.EqualError(t, err, "fake error trying to set sale")
			},
			wantSale: nil,
		},
		{
			name: "success",
			fields: fields{
				storage: NewLocalStorage(),
				apiURL:  mockServer.URL,
			},
			args: args{
				sale: &Sale{
					UserID: "1234",
					Amount: 100.0,
				},
			},
			wantErr: func(t *testing.T, err error) {
				require.Nil(t, err)
			},
			wantSale: func(t *testing.T, input *Sale) {
				require.NotEmpty(t, input.ID)
				require.NotEmpty(t, input.CreatedAt)
				require.NotEmpty(t, input.UpdatedAt)
				require.Equal(t, 1, input.Version)
			},
		},
		{
			name: "invalid amount",
			fields: fields{
				storage: NewLocalStorage(),
				apiURL:  mockServer.URL,
			},
			args: args{
				sale: &Sale{
					UserID: "1234",
					Amount: 0.0,
				},
			},
			wantErr: func(t *testing.T, err error) {
				require.NotNil(t, err)
				require.Equal(t, ErrInvalidInput, err)
			},
			wantSale: nil,
		},
		{
			name: "non-existent user",
			fields: fields{
				storage: NewLocalStorage(),
				apiURL:  mockServer.URL,
			},
			args: args{
				sale: &Sale{
					UserID: "9999",
					Amount: 100.0,
				},
			},
			wantErr: func(t *testing.T, err error) {
				require.NotNil(t, err)
				require.Equal(t, ErrUserNotFound, err)
			},
			wantSale: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(tt.fields.storage, nil, tt.fields.apiURL)

			err := s.CreateSale(tt.args.sale)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
			}
			if tt.wantSale != nil {
				tt.wantSale(t, tt.args.sale)
			}
		})
	}
}

func TestService_UpdateSale(t *testing.T) {
	type fields struct {
		storage Storage
	}

	type args struct {
		id      string
		updates *UpdateFieldsSale
	}

	tests := []struct {
		name      string
		fields    fields
		setupData func(storage Storage) string // para inicializar una venta y devolver su ID
		args      func(saleID string) args
		wantErr   func(t *testing.T, err error)
		wantSale  func(t *testing.T, sale *Sale)
	}{
		{
			name: "sale not found",
			fields: fields{
				storage: NewLocalStorage(),
			},
			setupData: func(_ Storage) string { return "no existe id" },
			args: func(id string) args {
				return args{id: id, updates: &UpdateFieldsSale{Status: "approved"}}
			},
			wantErr: func(t *testing.T, err error) {
				require.NotNil(t, err)
				require.Equal(t, ErrSaleNotFound, err)
			},
			wantSale: nil,
		},
		{
			name: "invalid status update",
			fields: fields{
				storage: NewLocalStorage(),
			},
			setupData: func(storage Storage) string {
				sale := &Sale{ID: "123", UserID: "1234", Amount: 100, Status: "pending"}
				_ = storage.SetSale(sale)
				return sale.ID
			},
			args: func(id string) args {
				return args{id: id, updates: &UpdateFieldsSale{Status: "invalid"}}
			},
			wantErr: func(t *testing.T, err error) {
				require.NotNil(t, err)
				require.Equal(t, ErrInvalidInput, err)
			},
			wantSale: nil,
		},
		{
			name: "invalid transaction - status not pending",
			fields: fields{
				storage: NewLocalStorage(),
			},
			setupData: func(storage Storage) string {
				sale := &Sale{ID: "456", UserID: "1234", Amount: 100, Status: "approved"}
				_ = storage.SetSale(sale)
				return sale.ID
			},
			args: func(id string) args {
				return args{id: id, updates: &UpdateFieldsSale{Status: "rejected"}}
			},
			wantErr: func(t *testing.T, err error) {
				require.NotNil(t, err)
				require.Equal(t, ErrTransactionInvalid, err)
			},
			wantSale: nil,
		},
		{
			name: "success",
			fields: fields{
				storage: NewLocalStorage(),
			},
			setupData: func(storage Storage) string {
				sale := &Sale{ID: "789", UserID: "1234", Amount: 100, Status: "pending", Version: 1}
				_ = storage.SetSale(sale)
				return sale.ID
			},
			args: func(id string) args {
				return args{id: id, updates: &UpdateFieldsSale{Status: "approved"}}
			},
			wantErr: func(t *testing.T, err error) {
				require.Nil(t, err)
			},
			wantSale: func(t *testing.T, sale *Sale) {
				require.Equal(t, "approved", sale.Status)
				require.Equal(t, 2, sale.Version)
				require.WithinDuration(t, time.Now(), sale.UpdatedAt, time.Second)
			},
		},
		{
			name: "error saving updated sale",
			fields: fields{
				storage: &mockStorage{
					mockReadSale: func(id string) (*Sale, error) {
						return &Sale{ID: id, Status: "pending", Version: 1}, nil
					},
					mockSetSale: func(sale *Sale) error {
						return errors.New("failed to save")
					},
				},
			},
			setupData: func(_ Storage) string { return "mock-id" },
			args: func(id string) args {
				return args{id: id, updates: &UpdateFieldsSale{Status: "approved"}}
			},
			wantErr: func(t *testing.T, err error) {
				require.NotNil(t, err)
				require.EqualError(t, err, "failed to save")
			},
			wantSale: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saleID := tt.setupData(tt.fields.storage)
			service := NewService(tt.fields.storage, nil, "")

			result, err := service.UpdateSale(tt.args(saleID).id, tt.args(saleID).updates)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
			}
			if tt.wantSale != nil {
				tt.wantSale(t, result)
			}
		})
	}
}

type mockStorage struct {
	mockSetSale      func(sale *Sale) error
	mockReadSale     func(id string) (*Sale, error)
	mockReadAllSales func() (map[string]*Sale, error)
}

func (m *mockStorage) SetSale(sale *Sale) error {
	return m.mockSetSale(sale)
}

func (m *mockStorage) ReadSale(id string) (*Sale, error) {
	return m.mockReadSale(id)
}

func (m *mockStorage) ReadAllSales() (map[string]*Sale, error) {
	return m.mockReadAllSales()
}
