package security_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/security"
)

func TestHashPassword_ReturnsNonEmptyHash(t *testing.T) {
	t.Parallel()

	hash, err := security.HashPassword("my-secret")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
}

func TestHashPassword_ProducesValidBcryptHash(t *testing.T) {
	t.Parallel()

	plain := "correct-horse-battery-staple"
	hash, err := security.HashPassword(plain)
	require.NoError(t, err)

	// Verify using bcrypt directly so we aren't just round-tripping our own code.
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)))
}

func TestCheckPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		givePlain string
		giveCheck string
		wantErr   bool
	}{
		{
			name:      "matching password returns nil",
			givePlain: "password123",
			giveCheck: "password123",
			wantErr:   false,
		},
		{
			name:      "wrong password returns error",
			givePlain: "password123",
			giveCheck: "wrongpassword",
			wantErr:   true,
		},
		{
			name:      "empty plain with non-empty hash returns error",
			givePlain: "password123",
			giveCheck: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := security.HashPassword(tt.givePlain)
			require.NoError(t, err)

			err = security.CheckPassword(hash, tt.giveCheck)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHashPassword_DifferentHashesEachCall(t *testing.T) {
	t.Parallel()

	hash1, err := security.HashPassword("same-password")
	require.NoError(t, err)

	hash2, err := security.HashPassword("same-password")
	require.NoError(t, err)

	// bcrypt uses a random salt, so two hashes of the same password differ.
	require.NotEqual(t, hash1, hash2)
}
