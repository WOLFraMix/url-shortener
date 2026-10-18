package handlers

import "testing"

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
