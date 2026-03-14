package utils_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/utils"
)

func TestGenerateID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		giveByteLen int
		wantHexLen  int
	}{
		{
			name:        "16 bytes produces 32-character hex",
			giveByteLen: 16,
			wantHexLen:  32,
		},
		{
			name:        "8 bytes produces 16-character hex",
			giveByteLen: 8,
			wantHexLen:  16,
		},
		{
			name:        "1 byte produces 2-character hex",
			giveByteLen: 1,
			wantHexLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := utils.GenerateID(tt.giveByteLen)
			require.NoError(t, err)
			require.Len(t, got, tt.wantHexLen)

			// Ensure the result is valid hex.
			_, decodeErr := hex.DecodeString(got)
			require.NoError(t, decodeErr)
		})
	}
}

func TestGenerateID_ProducesUniqueValues(t *testing.T) {
	t.Parallel()

	id1, err := utils.GenerateID(16)
	require.NoError(t, err)

	id2, err := utils.GenerateID(16)
	require.NoError(t, err)

	require.NotEqual(t, id1, id2)
}

func TestPtr(t *testing.T) {
	t.Parallel()

	t.Run("int pointer", func(t *testing.T) {
		t.Parallel()

		v := 42
		p := utils.Ptr(v)
		require.NotNil(t, p)
		require.Equal(t, v, *p)
	})

	t.Run("string pointer", func(t *testing.T) {
		t.Parallel()

		v := "hello"
		p := utils.Ptr(v)
		require.NotNil(t, p)
		require.Equal(t, v, *p)
	})

	t.Run("bool pointer", func(t *testing.T) {
		t.Parallel()

		v := true
		p := utils.Ptr(v)
		require.NotNil(t, p)
		require.Equal(t, v, *p)
	})

	t.Run("zero value pointer", func(t *testing.T) {
		t.Parallel()

		p := utils.Ptr(0)
		require.NotNil(t, p)
		require.Equal(t, 0, *p)
	})
}
