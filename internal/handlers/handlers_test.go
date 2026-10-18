package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string // человекочитаемое имя кейса для отчёта
		raw       string // вход: что скармливаем ValidateURL
		wantValid bool   // ожидание: true — пропускает, false — ошибка
	}{
		{name: "valid https", raw: "https://go.dev", wantValid: true},
		{name: "empty string", raw: "", wantValid: false},
		{name: "hello", raw: "hello", wantValid: false},
		{name: "invalid scheme", raw: "ftp://example.com", wantValid: false},
		{name: "empty host", raw: "http://", wantValid: false},
		{name: "cyrillic domain", raw: "https://пример.ру", wantValid: true},
		{name: "localhost with port", raw: "http://localhost:8080", wantValid: true},
		{name: "valid https with /", raw: "https://go.dev/", wantValid: true},
		{name: "only spaces", raw: "   ", wantValid: false},
		{name: "\n", raw: "\n", wantValid: false},
		{name: "int", raw: "12345", wantValid: false},
		{name: "quoted literal", raw: "''", wantValid: false},
		{name: "://.", raw: "://.", wantValid: false},
		{name: "dots", raw: "...", wantValid: false},
		{name: "https://12345.ru", raw: "https://12345.ru", wantValid: true},
		{name: "valid https in brackets", raw: "(https://go.dev)", wantValid: false},
	}

	for _, tc := range tests {
		err := ValidateURL(tc.raw)
		if tc.wantValid && err != nil {
			t.Errorf("%s: expected valid, got error: %v", tc.name, err)
		}
		if !tc.wantValid && err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}
}

// fakeStorage — хранилище в памяти для тестов хендлеров.
type fakeStorage struct {
	links     map[string]string // code -> originalURL
	errOnGet  error             // если задан — GetByShortCode вернёт эту ошибку
	errOnSave error             // если задан — Save вернёт эту ошибку
	savedURLs []string          // запись: что попало в Save
}

func (fs *fakeStorage) Save(originalURL, shortCode string) error {
	if fs.errOnSave != nil {
		return fs.errOnSave
	}
	if fs.links == nil {
		fs.links = map[string]string{}
	}
	fs.links[shortCode] = originalURL
	fs.savedURLs = append(fs.savedURLs, originalURL)
	return nil
}

func (fs *fakeStorage) GetByShortCode(code string) (string, error) {
	// ветка 1: ломалка включена — «симулируем» отказ базы
	if fs.errOnGet != nil {
		return "", fs.errOnGet
	}

	// ветка 2: нашли в map — обычный счастливый путь
	if url, ok := fs.links[code]; ok {
		return url, nil
	}

	// ветка 3: не нашли — говорим тем же языком, что Postgres
	return "", pgx.ErrNoRows
}

func (fs *fakeStorage) GetByURL(originalURL string) (string, error) {
	for code, url := range fs.links {
		if url == originalURL {
			return code, nil
		}
	}
	return "", pgx.ErrNoRows
}

func TestGetByShortCodeHandler(t *testing.T) {
	tests := []struct {
		name         string
		store        fakeStorage // готовое «состояние базы»
		code         string      // путь, который дёргаем
		wantStatus   int
		wantLocation string // ожидаемый Location (или "")
	}{
		{
			name:         "redirects to original",
			store:        fakeStorage{links: map[string]string{"abc12345": "https://go.dev"}},
			code:         "abc12345",
			wantStatus:   http.StatusFound,
			wantLocation: "https://go.dev",
		},
		{
			name:         "unknown code",
			store:        fakeStorage{}, // пустая база
			code:         "missing",
			wantStatus:   http.StatusNotFound,
			wantLocation: "",
		},
		{
			name:         "storage error",
			store:        fakeStorage{errOnGet: errors.New("db down")},
			code:         "abc12345",
			wantStatus:   http.StatusInternalServerError,
			wantLocation: "",
		},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, "/"+tc.code, nil)
		req.SetPathValue("code", tc.code)

		rec := httptest.NewRecorder()

		GetByShortCodeHandler(&tc.store)(rec, req)

		if rec.Code != tc.wantStatus {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.wantStatus)
		}
		if loc := rec.Header().Get("Location"); loc != tc.wantLocation {
			t.Errorf("%s: Location = %q, want %q", tc.name, loc, tc.wantLocation)
		}
	}
}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string // сырое тело запроса
		store      fakeStorage
		wantStatus int
		wantCode   string // ожидаемый short_code ("" — не проверяем)
	}{
		{
			name:       "valid url",
			body:       `{"url": "https://go.dev"}`,
			store:      fakeStorage{links: map[string]string{}},
			wantStatus: http.StatusCreated,
			wantCode:   "bn9Y9rho",
		},
		{
			name:       "invalid json",
			body:       `not json at all`,
			store:      fakeStorage{links: map[string]string{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid url",
			body:       `{"url": "hello"}`,
			store:      fakeStorage{links: map[string]string{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate url",
			body: `{"url": "https://go.dev"}`,
			store: fakeStorage{
				links:     map[string]string{"bn9Y9rho": "https://go.dev"},
				errOnSave: &pgconn.PgError{Code: "23505"}, // «база» говорит: дубликат
			},
			wantStatus: http.StatusCreated,
			wantCode:   "bn9Y9rho", // существующий код, а не новая генерация
		},
		{
			name:       "save failure",
			body:       `{"url": "https://go.dev"}`,
			store:      fakeStorage{links: map[string]string{}, errOnSave: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tc.body))
		rec := httptest.NewRecorder()

		ShortenHandler(&tc.store)(rec, req)

		if rec.Code != tc.wantStatus {
			t.Errorf("%s: status = %d, want %d", tc.name, rec.Code, tc.wantStatus)
		}

		// разбор ответа: тело обязано быть корректным JSON во всех кейсах
		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("%s: response is not valid json: %v", tc.name, err)
		}

		if tc.wantCode != "" && resp["short_code"] != tc.wantCode {
			t.Errorf("%s: short_code = %q, want %q", tc.name, resp["short_code"], tc.wantCode)
		}

		// шпион: для успешных кейсов Save должен был получить ровно нашу ссылку
		if tc.wantCode != "" && tc.name != "duplicate url" {
			if len(tc.store.savedURLs) != 1 {
				t.Errorf("%s: saved %d urls, want 1", tc.name, len(tc.store.savedURLs))
			} else if tc.store.savedURLs[0] != "https://go.dev" {
				t.Errorf("%s: saved url = %q, want %q", tc.name, tc.store.savedURLs[0], "https://go.dev")
			}
		}
	}
}
