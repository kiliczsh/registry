package v0_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v0 "github.com/modelcontextprotocol/registry/internal/api/handlers/v0"
	"github.com/modelcontextprotocol/registry/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestSeedHandler(t *testing.T) {
	t.Run("successful seed export", func(t *testing.T) {
		mockRegistry := new(MockRegistryService)

		// Mock data
		servers := []model.Server{
			{
				ID:          "test-id-1",
				Name:        "test-server-1",
				Description: "Test server 1",
				Repository: model.Repository{
					URL:    "https://github.com/test/repo1",
					Source: "github",
					ID:     "123",
				},
				VersionDetail: model.VersionDetail{
					Version:     "1.0.0",
					ReleaseDate: "2023-01-01T00:00:00Z",
					IsLatest:    true,
				},
			},
		}

		serverDetail := &model.ServerDetail{
			Server: servers[0],
			Packages: []model.Package{
				{
					RegistryName: "npm",
					Name:         "test-package",
					Version:      "1.0.0",
				},
			},
		}

		mockRegistry.Mock.On("List", "", 10000).Return(servers, "", nil)
		mockRegistry.Mock.On("GetByID", "test-id-1").Return(serverDetail, nil)

		handler := v0.SeedHandler(mockRegistry)

		req, err := http.NewRequest("GET", "/v0/seed.json", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response []model.ServerDetail
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, "test-server-1", response[0].Name)
		assert.Len(t, response[0].Packages, 1)

		mockRegistry.Mock.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockRegistry := new(MockRegistryService)
		handler := v0.SeedHandler(mockRegistry)

		req, err := http.NewRequest("POST", "/v0/seed.json", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		assert.Equal(t, "GET", rr.Header().Get("Allow"))
	})

	t.Run("registry service error", func(t *testing.T) {
		mockRegistry := new(MockRegistryService)
		mockRegistry.Mock.On("List", "", 10000).Return([]model.Server{}, "", assert.AnError)

		handler := v0.SeedHandler(mockRegistry)

		req, err := http.NewRequest("GET", "/v0/seed.json", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockRegistry.Mock.AssertExpectations(t)
	})
}
