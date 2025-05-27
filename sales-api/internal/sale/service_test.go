package sale

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
	mockHandler.HandleFunc("/users/1234", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"1234"}`)) // JSON válido
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
			name: "error",
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
