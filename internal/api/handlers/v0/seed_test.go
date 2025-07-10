package v0

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/registry/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRegistryService is a mock implementation of RegistryService for testing
type MockRegistryService struct {
	mock.Mock
}

func (m *MockRegistryService) List(cursor string, limit int) ([]model.Server, string, error) {
	args := m.Called(cursor, limit)
	return args.Get(0).([]model.Server), args.String(1), args.Error(2)
}

func (m *MockRegistryService) GetByID(id string) (*model.ServerDetail, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ServerDetail), args.Error(1)
}

func (m *MockRegistryService) Publish(serverDetail *model.ServerDetail) error {
	args := m.Called(serverDetail)
	return args.Error(0)
}

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

		mockRegistry.On("List", "", 10000).Return(servers, "", nil)
		mockRegistry.On("GetByID", "test-id-1").Return(serverDetail, nil)

		handler := SeedHandler(mockRegistry)

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

		mockRegistry.AssertExpectations(t)
	})

	t.Run("method not allowed", func(t *testing.T) {
		mockRegistry := new(MockRegistryService)
		handler := SeedHandler(mockRegistry)

		req, err := http.NewRequest("POST", "/v0/seed.json", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		assert.Equal(t, "GET", rr.Header().Get("Allow"))
	})

	t.Run("registry service error", func(t *testing.T) {
		mockRegistry := new(MockRegistryService)
		mockRegistry.On("List", "", 10000).Return([]model.Server{}, "", assert.AnError)

		handler := SeedHandler(mockRegistry)

		req, err := http.NewRequest("GET", "/v0/seed.json", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockRegistry.AssertExpectations(t)
	})
}
