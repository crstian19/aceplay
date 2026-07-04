// Package acestream provides functionality for parsing AceStream URLs
package acestream

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	// ErrInvalidURL is returned when the URL is invalid
	ErrInvalidURL = errors.New("invalid acestream URL")

	// ErrEmptyContentID is returned when the content ID is empty
	ErrEmptyContentID = errors.New("empty content ID")

	// acestreamRegex is the regex pattern to validate acestream:// URLs
	// Supports format: acestream://[content_id]
	acestreamRegex = regexp.MustCompile(`^acestream://([a-fA-F0-9]{40})$`)
)

// URL represents a parsed AceStream URL
type URL struct {
	// ContentID is the unique content identifier (40-character hex hash)
	ContentID string

	// Raw is the original unprocessed URL
	Raw string
}

// ParseURL parses an acestream URL and returns a URL structure
// Supports formats:
//   - acestream://[content_id]
//   - acestream://[content_id]?options
//
// Example:
//
//	url, err := ParseURL("acestream://abcd1234...")
func ParseURL(rawURL string) (*URL, error) {
	if strings.TrimSpace(rawURL) == "" {
		return nil, ErrInvalidURL
	}

	// Normalize the URL
	rawURL = strings.TrimSpace(rawURL)

	// Try to parse as standard URL first
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		// If it fails, try with regex
		return parseWithRegex(rawURL)
	}

	// Verify the scheme is acestream
	if parsedURL.Scheme != "acestream" {
		return nil, fmt.Errorf("%w: scheme must be 'acestream', got '%s'", ErrInvalidURL, parsedURL.Scheme)
	}

	// Extract the content ID (it's in the host or path)
	contentID := parsedURL.Host
	if contentID == "" {
		contentID = strings.TrimPrefix(parsedURL.Path, "/")
	}

	if err := validateContentID(contentID); err != nil {
		return nil, err
	}

	return &URL{
		ContentID: strings.ToLower(contentID),
		Raw:       rawURL,
	}, nil
}

// parseWithRegex tries to parse using regex as fallback
func parseWithRegex(rawURL string) (*URL, error) {
	matches := acestreamRegex.FindStringSubmatch(rawURL)
	if len(matches) != 2 {
		return nil, fmt.Errorf("%w: invalid format, must be acestream://[40 hex characters]", ErrInvalidURL)
	}

	contentID := strings.ToLower(matches[1])

	if err := validateContentID(contentID); err != nil {
		return nil, err
	}

	return &URL{
		ContentID: contentID,
		Raw:       rawURL,
	}, nil
}

// validateContentID validates that the content ID is valid
func validateContentID(contentID string) error {
	if contentID == "" {
		return ErrEmptyContentID
	}

	// Verify length (must be 40 hex characters)
	if len(contentID) != 40 {
		return fmt.Errorf("%w: content ID must have 40 hexadecimal characters, has %d", ErrInvalidURL, len(contentID))
	}

	// Verify it's hexadecimal
	if !isHexString(contentID) {
		return fmt.Errorf("%w: content ID must contain only hexadecimal characters (0-9, a-f)", ErrInvalidURL)
	}

	return nil
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

// String returns the string representation of the URL
func (u *URL) String() string {
	return fmt.Sprintf("acestream://%s", u.ContentID)
}

// IsValid checks if the URL is valid without fully parsing it
func IsValid(rawURL string) bool {
	_, err := ParseURL(rawURL)
	return err == nil
}

// ExtractContentID extracts only the content ID from a URL without full validation
func ExtractContentID(rawURL string) (string, error) {
	url, err := ParseURL(rawURL)
	if err != nil {
		return "", err
	}
	return url.ContentID, nil
}
