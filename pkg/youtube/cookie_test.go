package youtube

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestLoadCookiesFromFile(t *testing.T) {
	t.Run("valid netscape format", func(t *testing.T) {
		content := `# Netscape HTTP Cookie File
# This is a generated file!  Do not edit.

.youtube.com	TRUE	/	TRUE	1735689600	__Secure-1PSID	test_value_1
.youtube.com	TRUE	/	FALSE	0	PREF	tz=UTC
.youtube.com	TRUE	/	TRUE	1735689600	__Secure-3PAPISID	test_value_2
`
		tmpfile := createTempCookieFile(t, content)
		defer func() { _ = os.Remove(tmpfile) }()

		cookies, err := LoadCookiesFromFile(tmpfile)
		if err != nil {
			t.Fatalf("LoadCookiesFromFile failed: %v", err)
		}

		if len(cookies) != 3 {
			t.Errorf("expected 3 cookies, got %d", len(cookies))
		}

		// Check first cookie
		if cookies[0].Name != "__Secure-1PSID" {
			t.Errorf("expected cookie name '__Secure-1PSID', got '%s'", cookies[0].Name)
		}
		if cookies[0].Value != "test_value_1" {
			t.Errorf("expected cookie value 'test_value_1', got '%s'", cookies[0].Value)
		}
		if cookies[0].Domain != ".youtube.com" {
			t.Errorf("expected domain '.youtube.com', got '%s'", cookies[0].Domain)
		}
		if !cookies[0].Secure {
			t.Error("expected cookie to be secure")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		content := `# Netscape HTTP Cookie File
# This is a generated file!  Do not edit.
`
		tmpfile := createTempCookieFile(t, content)
		defer func() { _ = os.Remove(tmpfile) }()

		cookies, err := LoadCookiesFromFile(tmpfile)
		if err != nil {
			t.Fatalf("LoadCookiesFromFile failed: %v", err)
		}

		if len(cookies) != 0 {
			t.Errorf("expected 0 cookies, got %d", len(cookies))
		}
	})

	t.Run("file with comments and blank lines", func(t *testing.T) {
		content := `# Netscape HTTP Cookie File
# This is a generated file!  Do not edit.

# This is a comment
.youtube.com	TRUE	/	TRUE	1735689600	COOKIE1	value1

# Another comment

.youtube.com	TRUE	/	FALSE	0	COOKIE2	value2
`
		tmpfile := createTempCookieFile(t, content)
		defer func() { _ = os.Remove(tmpfile) }()

		cookies, err := LoadCookiesFromFile(tmpfile)
		if err != nil {
			t.Fatalf("LoadCookiesFromFile failed: %v", err)
		}

		if len(cookies) != 2 {
			t.Errorf("expected 2 cookies, got %d", len(cookies))
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadCookiesFromFile("/nonexistent/file.txt")
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})

	t.Run("session cookie with expiry 0", func(t *testing.T) {
		content := `# Netscape HTTP Cookie File
.youtube.com	TRUE	/	FALSE	0	SESSION_COOKIE	session_value
`
		tmpfile := createTempCookieFile(t, content)
		defer func() { _ = os.Remove(tmpfile) }()

		cookies, err := LoadCookiesFromFile(tmpfile)
		if err != nil {
			t.Fatalf("LoadCookiesFromFile failed: %v", err)
		}

		if len(cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(cookies))
		}

		// Session cookies have no expiry set (zero time)
		if !cookies[0].Expires.IsZero() && cookies[0].Expires.Before(time.Now().Add(100*365*24*time.Hour)) {
			t.Error("session cookie should have no expiry or far future expiry")
		}
	})
}

func TestFilterSecureCookies(t *testing.T) {
	cookies := []*http.Cookie{
		{Name: "__Secure-1PSID", Value: "val1", Domain: ".youtube.com", Secure: true},
		{Name: "PREF", Value: "val2", Domain: ".youtube.com", Secure: false},
		{Name: "__Secure-3PAPISID", Value: "val3", Domain: ".youtube.com", Secure: true},
		{Name: "VISITOR_INFO1_LIVE", Value: "val4", Domain: ".youtube.com", Secure: false},
	}

	secure := FilterSecureCookies(cookies)

	if len(secure) != 2 {
		t.Errorf("expected 2 secure cookies, got %d", len(secure))
	}

	for _, c := range secure {
		if !c.Secure {
			t.Errorf("cookie %s should be secure", c.Name)
		}
		if c.Name != "__Secure-1PSID" && c.Name != "__Secure-3PAPISID" {
			t.Errorf("unexpected secure cookie: %s", c.Name)
		}
	}
}

// Helper function to create a temporary cookie file
func createTempCookieFile(t *testing.T, content string) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "cookies-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = tmpfile.Close() }()

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	return tmpfile.Name()
}

func TestParseCookieLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantCookie  *Cookie
		wantErr     bool
		description string
	}{
		{
			name:        "valid cookie",
			line:        ".youtube.com\tTRUE\t/\tTRUE\t1735689600\tCOOKIE_NAME\tcookie_value",
			wantCookie:  &Cookie{Name: "COOKIE_NAME", Value: "cookie_value", Domain: ".youtube.com", Path: "/", Secure: true},
			wantErr:     false,
			description: "standard valid cookie line",
		},
		{
			name:        "comment line",
			line:        "# This is a comment",
			wantCookie:  nil,
			wantErr:     false,
			description: "comment should be skipped",
		},
		{
			name:        "empty line",
			line:        "",
			wantCookie:  nil,
			wantErr:     false,
			description: "empty line should be skipped",
		},
		{
			name:        "insufficient fields",
			line:        ".youtube.com\tTRUE\t/",
			wantCookie:  nil,
			wantErr:     true,
			description: "line with insufficient fields should error",
		},
		{
			name:        "session cookie",
			line:        ".youtube.com\tTRUE\t/\tFALSE\t0\tSESSION\tsession_val",
			wantCookie:  &Cookie{Name: "SESSION", Value: "session_val", Domain: ".youtube.com", Path: "/", Secure: false},
			wantErr:     false,
			description: "session cookie with expiry 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCookieLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCookieLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got == nil && tt.wantCookie == nil {
				return
			}
			if got == nil || tt.wantCookie == nil {
				t.Errorf("parseCookieLine() got = %v, want %v", got, tt.wantCookie)
				return
			}
			if got.Name != tt.wantCookie.Name || got.Value != tt.wantCookie.Value {
				t.Errorf("parseCookieLine() got = %v, want %v", got, tt.wantCookie)
			}
		})
	}
}
