package orm

import "testing"

func TestToSnakeCase(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"ID", "id"},
		{"UserID", "user_id"},
		{"HTTP", "http"},
		{"HTTPRequest", "http_request"},
		{"HTTPServer", "http_server"},
		{"GetHTTPResponse", "get_http_response"},
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"URLParser", "url_parser"},
		{"XMLParser", "xml_parser"},
		{"ParseXML", "parse_xml"},
		{"HTTPS", "https"},
		{"HTTPSClient", "https_client"},
		{"SimpleHTTPServer", "simple_http_server"},
		{"User", "user"},
		{"CreatedAt", "created_at"},
	}

	for _, tc := range testCases {
		result := toSnakeCase(tc.input)
		if result != tc.expected {
			t.Errorf("toSnakeCase(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
