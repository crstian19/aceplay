package acestream

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContent string
		wantErr     bool
		errType     error
	}{
		{
			name:        "valid simple URL",
			input:       "acestream://abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantContent: "abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantErr:     false,
		},
		{
			name:        "valid URL with uppercase",
			input:       "acestream://ABCD1234ABCD1234ABCD1234ABCD1234ABCD1234",
			wantContent: "abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantErr:     false,
		},
		{
			name:        "valid URL with mixed hex",
			input:       "acestream://aBcD1234aBcD1234aBcD1234aBcD1234aBcD1234",
			wantContent: "abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantErr:     false,
		},
		{
			name:    "empty URL",
			input:   "",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "URL without scheme",
			input:   "abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "URL with wrong scheme",
			input:   "http://abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "content ID too short",
			input:   "acestream://abcd1234",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "content ID too long",
			input:   "acestream://abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:    "content ID with invalid characters",
			input:   "acestream://abcd1234abcd1234abcd1234abcd1234abcdzzzz",
			wantErr: true,
			errType: ErrInvalidURL,
		},
		{
			name:        "URL with surrounding spaces",
			input:       "  acestream://abcd1234abcd1234abcd1234abcd1234abcd1234  ",
			wantContent: "abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, got.ContentID)
		})
	}
}

func TestURL_String(t *testing.T) {
	url := &URL{
		ContentID: "abcd1234abcd1234abcd1234abcd1234abcd1234",
		Raw:       "acestream://abcd1234abcd1234abcd1234abcd1234abcd1234",
	}

	assert.Equal(t, "acestream://abcd1234abcd1234abcd1234abcd1234abcd1234", url.String())
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "acestream://abcd1234abcd1234abcd1234abcd1234abcd1234", true},
		{"empty", "", false},
		{"no scheme", "abcd1234abcd1234abcd1234abcd1234abcd1234", false},
		{"wrong scheme", "http://abcd1234abcd1234abcd1234abcd1234abcd1234", false},
		{"short content id", "acestream://abcd1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValid(tt.input))
		})
	}
}

func TestExtractContentID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid URL",
			input:   "acestream://abcd1234abcd1234abcd1234abcd1234abcd1234",
			want:    "abcd1234abcd1234abcd1234abcd1234abcd1234",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractContentID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateContentID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid 40 chars hex", "abcd1234abcd1234abcd1234abcd1234abcd1234", false},
		{"valid uppercase", "ABCD1234ABCD1234ABCD1234ABCD1234ABCD1234", false},
		{"valid numbers", "1234567890123456789012345678901234567890", false},
		{"empty", "", true},
		{"short 39 chars", "abcd1234abcd1234abcd1234abcd1234abcd123", true},
		{"long 41 chars", "abcd1234abcd1234abcd1234abcd1234abcd12345", true},
		{"invalid characters", "abcd1234abcd1234abcd1234abcd1234abcdzzzz", true},
		{"with dashes", "abcd1234-abcd1234-abcd1234-abcd1234-abcd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContentID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
