package sale

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Create_Simple(t *testing.T) {
	s := NewService(NewLocalStorage(), nil)

	input := &Sale{
		UserID: "1",
		Amount: 100.0,
	}

	var err error
	require.Nil(t, err)                  //valida que el error sea nil
	require.NotEmpty(t, input.ID)        //para validar que el ID no sea vacío
	require.NotEmpty(t, input.CreatedAt) //valida que la fecha de creación no sea vacía
	require.NotEmpty(t, input.UpdatedAt) //valida que la fecha de actualización no sea vacía
	require.Equal(t, 1, input.Version)   //valida que la versión sea 1

	s = NewService(&mockStorage{
		mockSetSale: func(sale *Sale) error {
			return errors.New("fake error trying to set sale")
		},
	}, nil)

	err = s.CreateSale(input)
	require.NotNil(t, err)
	require.EqualError(t, err, "fake error trying to set sale")
}

func TestService_Create(t *testing.T) {
	type fields struct {
		storage Storage
	}

	type args struct {
		sale *Sale
	}

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
			},
			args: args{
				sale: &Sale{},
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
			},
			args: args{
				sale: &Sale{
					UserID: "1",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				storage: tt.fields.storage,
			}

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
