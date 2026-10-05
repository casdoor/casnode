// Copyright 2026 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGoogleTranslate(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantText   string
		wantLang   string
		wantErrMsg string
	}{
		{
			name:     "success",
			status:   http.StatusOK,
			body:     `{"data":{"translations":[{"translatedText":"hola","detectedSourceLanguage":"en"}]}}`,
			wantText: "hola",
			wantLang: "en",
		},
		{
			name:       "api error",
			status:     http.StatusBadRequest,
			body:       `{"error":{"code":400,"message":"API key not valid"}}`,
			wantErrMsg: "API key not valid",
		},
		{
			name:       "empty translations",
			status:     http.StatusOK,
			body:       `{"data":{"translations":[]}}`,
			wantErrMsg: "Translate Failed",
		},
		{
			name:       "invalid json",
			status:     http.StatusBadGateway,
			body:       `<html>bad gateway</html>`,
			wantErrMsg: "invalid translation response (HTTP 502): invalid character '<' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Fatalf("ParseForm() error: %v", err)
				}
				if got := r.PostForm.Get("target"); got != "es" {
					t.Errorf("target = %q, want %q", got, "es")
				}
				if got := r.PostForm.Get("key"); got != "test-key" {
					t.Errorf("key = %q, want %q", got, "test-key")
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			text, lang, err := googleTranslate(server.URL, "test-key", "hello", "es")
			if tt.wantErrMsg != "" {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Fatalf("googleTranslate() error = %v, want %q", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("googleTranslate() unexpected error: %v", err)
			}
			if text != tt.wantText || lang != tt.wantLang {
				t.Fatalf("googleTranslate() = (%q, %q), want (%q, %q)", text, lang, tt.wantText, tt.wantLang)
			}
		})
	}
}

func TestGoogleTranslateRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	endpoint := server.URL
	server.Close()

	if _, _, err := googleTranslate(endpoint, "test-key", "hello", "es"); err == nil {
		t.Fatal("googleTranslate() expected an error for an unreachable endpoint")
	}
}
