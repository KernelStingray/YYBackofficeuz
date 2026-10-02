package sanitize

import "html"

// String escapes HTML special characters to prevent XSS attacks.
func String(s string) string {
	return html.EscapeString(s)
}

// StringPtr escapes HTML special characters if the pointer is not nil.
func StringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	escaped := html.EscapeString(*s)
	return &escaped
}
