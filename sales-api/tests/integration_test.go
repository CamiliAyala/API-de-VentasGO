package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sales-api/api"
	"sales-api/internal/sale"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestIntegrationCreateAndGet(t *testing.T) {
	app := gin.Default()
	api.InitRoutes(app)

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	res := fakeRequest(app, req)

	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "pong")

	req, _ = http.NewRequest(http.MethodPost, "/sales", bytes.NewBufferString(`{
		"user_id":"1",
		"amount": 100.0	
	}`))

	res = fakeRequest(app, req)

	require.NotNil(t, res)
	require.Equal(t, http.StatusCreated, res.Code)

	var resSale *sale.Sale
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &resSale))
	require.Equal(t, "1", resSale.UserID) //no sabemos q ID es el correcto cuando ejecuta
	require.Equal(t, 100.0, resSale.Amount)
	require.Equal(t, 1, resSale.Version)
	require.NotEmpty(t, resSale.Status)
	require.NotEmpty(t, resSale.ID)
	require.NotEmpty(t, resSale.CreatedAt)
	require.NotEmpty(t, resSale.UpdatedAt)

	req, _ = http.NewRequest(http.MethodGet, "/sales?user_id="+resSale.UserID, nil)

	res = fakeRequest(app, req)

	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)

	req, _ = http.NewRequest(http.MethodGet, "/sales?user_id="+resSale.UserID+"&status="+resSale.Status, nil)

	res = fakeRequest(app, req)

	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
}

func fakeRequest(e *gin.Engine, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	return w
}
