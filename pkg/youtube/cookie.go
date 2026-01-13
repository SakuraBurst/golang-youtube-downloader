package youtube

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Cookie represents an HTTP cookie with YouTube-specific fields.
type Cookie struct {
	Name     string
	Value    string
	Domain   string
	Path     string
	Expires  time.Time
	Secure   bool
	HttpOnly bool
}

// ToHTTPCookie converts a Cookie to http.Cookie format.
func (c *Cookie) ToHTTPCookie() *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Domain:   c.Domain,
		Path:     c.Path,
		Expires:  c.Expires,
		Secure:   c.Secure,
		HttpOnly: c.HttpOnly,
	}
}

// LoadCookiesFromFile loads cookies from a Netscape format cookie file.
// The Netscape format is a tab-separated format used by browsers and tools like yt-dlp.
//
// Format:
// domain	flag	path	secure	expiration	name	value
//
// Example:
// .youtube.com	TRUE	/	TRUE	1735689600	__Secure-1PSID	test_value
func LoadCookiesFromFile(filename string) ([]*http.Cookie, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening cookie file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	var cookies []*http.Cookie
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		cookie, err := parseCookieLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		if cookie != nil {
			cookies = append(cookies, cookie.ToHTTPCookie())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading cookie file: %w", err)
	}

	return cookies, nil
}

// parseCookieLine parses a single line from a Netscape cookie file.
// Returns nil, nil for comment or empty lines.
func parseCookieLine(line string) (*Cookie, error) {
	line = strings.TrimSpace(line)

	// Skip comments and empty lines
	if line == "" || strings.HasPrefix(line, "#") {
		return nil, nil
	}

	// Netscape format: domain	flag	path	secure	expiration	name	value
	parts := strings.Split(line, "\t")
	if len(parts) < 7 {
		return nil, fmt.Errorf("invalid cookie format: expected 7 fields, got %d", len(parts))
	}

	// Parse expiration time
	expirationStr := strings.TrimSpace(parts[4])
	var expires time.Time
	if expirationStr != "0" {
		expirationUnix, err := strconv.ParseInt(expirationStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing expiration time: %w", err)
		}
		expires = time.Unix(expirationUnix, 0)
	} else {
		// Session cookie (expiry 0) - set to far future to avoid expiration
		expires = time.Now().Add(100 * 365 * 24 * time.Hour)
	}

	// Parse secure flag
	secureStr := strings.TrimSpace(parts[3])
	secure := strings.EqualFold(secureStr, "TRUE")

	cookie := &Cookie{
		Domain:  strings.TrimSpace(parts[0]),
		Path:    strings.TrimSpace(parts[2]),
		Secure:  secure,
		Expires: expires,
		Name:    strings.TrimSpace(parts[5]),
		Value:   strings.TrimSpace(parts[6]),
	}

	return cookie, nil
}

// FilterSecureCookies filters cookies to return only those marked as secure.
// This is useful for identifying YouTube authentication cookies which are
// typically prefixed with __Secure-.
func FilterSecureCookies(cookies []*http.Cookie) []*http.Cookie {
	var secure []*http.Cookie
	for _, c := range cookies {
		if c.Secure {
			secure = append(secure, c)
		}
	}
	return secure
}

// IsAuthenticated checks if the provided cookies contain valid YouTube
// authentication cookies. It verifies that __Secure cookies exist and are not expired.
func IsAuthenticated(cookies []*http.Cookie) bool {
	now := time.Now()
	hasSecureCookie := false

	for _, c := range cookies {
		// Check for __Secure cookies (YouTube auth cookies)
		if strings.HasPrefix(c.Name, "__Secure-") {
			hasSecureCookie = true
			// Check if cookie is expired
			if !c.Expires.IsZero() && c.Expires.Before(now) {
				return false
			}
		}
	}

	return hasSecureCookie
}
