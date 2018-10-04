package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vikramsk/rediproxy/pkg/cache"
	"github.com/vikramsk/rediproxy/pkg/internal/mocks"
)

type scenario struct {
	name           string
	reqURL         string
	expectedStatus int
	proxyService   *mocks.Getter
}

func cacheHit(key string) (string, error) {
	return key, nil
}

func cacheMiss(key string) (string, error) {
	return "", cache.ErrKeyNotFound
}

func internalError(key string) (string, error) {
	return "", errors.New("internal error")
}

func TestInvalidRequest(t *testing.T) {
	scenarios := []scenario{
		{
			name:           "invalid request route should return not found",
			reqURL:         "http://test/invalid",
			expectedStatus: http.StatusNotFound,
			proxyService:   nil,
		},
		{
			name:           "empty key should return bad request",
			reqURL:         "http://test/cache?key=",
			expectedStatus: http.StatusBadRequest,
			proxyService:   nil,
		},
		{
			name:           "key not found error should return no content",
			reqURL:         "http://test/cache?key=test",
			expectedStatus: http.StatusNoContent,
			proxyService: &mocks.Getter{
				GetFn: cacheMiss,
			},
		},
		{
			name:           "service error should return internal server error",
			reqURL:         "http://test/cache?key=test",
			expectedStatus: http.StatusInternalServerError,
			proxyService: &mocks.Getter{
				GetFn: internalError,
			},
		},
		{
			name:           "valid output with no errors should return OK",
			reqURL:         "http://test/cache?key=test",
			expectedStatus: http.StatusOK,
			proxyService: &mocks.Getter{
				GetFn: cacheHit,
			},
		},
	}

	for i := range scenarios {
		req := httptest.NewRequest("GET", scenarios[i].reqURL, nil)
		w := httptest.NewRecorder()
		handler := NewProxyHandler(scenarios[i].proxyService)
		handler.ServeHTTP(w, req)

		resp := w.Result()

		if resp.StatusCode != scenarios[i].expectedStatus {
			t.Errorf("API Handler test failed for: %s, expected: %d, received: %d", scenarios[i].name, scenarios[i].expectedStatus, resp.StatusCode)
		}
	}
}
