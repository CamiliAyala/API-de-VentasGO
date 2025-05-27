package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sales-api/api"
	"sales-api/internal/sale"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestIntegrationCreateAndPatchAndGet(t *testing.T) {
	mockHandler := http.NewServeMux()

	mockHandler.HandleFunc("/users/1234", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`1234`))
	})

	mockServer := httptest.NewServer(mockHandler)
	defer mockServer.Close()

	app := gin.Default()
	api.InitRoutes(app, mockServer.URL)

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	res := fakeRequest(app, req)

	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "pong")

	//flujo completo de POST → PATCH → GET en el happy path.

	req, _ = http.NewRequest(http.MethodPost, "/sales", bytes.NewBufferString(`{
		"user_id": "1234",
		"amount": 100.0	
	}`))

	res = fakeRequest(app, req)
	var amount1 float32 = 100.0
	require.NotNil(t, res)
	require.Equal(t, http.StatusCreated, res.Code)

	var resSale *sale.Sale
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &resSale))
	require.Equal(t, "1234", resSale.UserID)
	require.Equal(t, amount1, resSale.Amount)
	require.Equal(t, 1, resSale.Version)
	require.NotEmpty(t, resSale.Status)
	require.NotEmpty(t, resSale.ID)
	require.NotEmpty(t, resSale.CreatedAt)
	require.NotEmpty(t, resSale.UpdatedAt)

	statusActual := resSale.Status
	req, _ = http.NewRequest(http.MethodPatch, "/sales/"+resSale.ID, bytes.NewBufferString(`{
		"status":"approved"
	}`))
	res = fakeRequest(app, req)

	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &resSale))

	if statusActual == "pending" {
		require.NotNil(t, res)
		require.Equal(t, "approved", resSale.Status)
		require.Equal(t, 2, resSale.Version)
		require.WithinDuration(t, time.Now(), resSale.UpdatedAt, time.Second)
		require.Equal(t, http.StatusOK, res.Code)
	} else {
		require.NotNil(t, res)
		require.Equal(t, http.StatusConflict, res.Code)
		require.Contains(t, res.Body.String(), "transaccion invalida")
	}

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
