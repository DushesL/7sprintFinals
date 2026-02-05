package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		c int
		w int
	}{
		{c: 0, w: 0},
		{c: 1, w: 1},
		{c: 2, w: 2},
		{c: 100, w: len(cafeList["moscow"])},
	}

	for _, tt := range requests {
		url := "/cafe?city=moscow&count=" + strconv.Itoa(tt.c)

		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		// Проверяем, что запрос обработан успешно
		require.Equal(t, http.StatusOK, response.Code, "запрос с count=%d должен возвращать 200 OK", tt.c)

		body := strings.TrimSpace(response.Body.String())
		var cafeNames []string
		if body != "" {
			cafeNames = strings.Split(body, ",")
		}

		assert.Equal(t, tt.w, len(cafeNames), "для count=%d ожидаем %d кафе, получено %d", tt.c, tt.w, len(cafeNames))
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search string
		wc     int
	}{
		{search: "фасоль", wc: 0},
		{search: "кофе", wc: 2},
		{search: "вилка", wc: 1},
	}

	for _, tt := range requests {
		url := "/cafe?city=moscow&search=" + tt.search

		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code, "request with search=%q must retur 200 OK", tt.search)

		body := strings.TrimSpace(response.Body.String())
		var cafeNames []string
		if body != "" {
			cafeNames = strings.Split(body, ",")
		}

		assert.Equal(t, tt.wc, len(cafeNames), "for search=%q wait %d cafe, got %d", tt.search, tt.wc, len(cafeNames))

		searchLower := strings.ToLower(tt.search)
		for _, name := range cafeNames {
			nameLower := strings.ToLower(name)
			assert.True(t, strings.Contains(nameLower, searchLower),
				"cade name %q must include %q (case insensetive)", name, tt.search)
		}
	}
}
