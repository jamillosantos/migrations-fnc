package migrationsfnc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getMigrationInfo(t *testing.T) {
	tests := []struct {
		name            string
		file            string
		wantID          string
		wantDescription string
		wantErr         error
	}{
		{"valid", "1234567890_some_description.go", "1234567890", "some description", nil},
		{"no description", "1234567890.go", "1234567890", "", ErrMigrationDescriptionRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getMigrationInfo(tt.file)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantID, got)
			assert.Equal(t, tt.wantDescription, got1)
		})
	}
}
