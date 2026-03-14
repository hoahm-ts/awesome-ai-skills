package shared_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hoahm-ts/awesome-ai-skills/internal/shared"
)

func TestSentinelErrors_AreDistinct(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		shared.ErrNotFound,
		shared.ErrAlreadyExists,
		shared.ErrUnauthorized,
		shared.ErrForbidden,
		shared.ErrInvalidInput,
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i == j {
				continue
			}
			require.False(t, errors.Is(a, b), "expected %v and %v to be distinct", a, b)
		}
	}
}

func TestSentinelErrors_MatchWithErrorsIs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		giveErr error
		wantErr error
	}{
		{name: "ErrNotFound matches itself", giveErr: shared.ErrNotFound, wantErr: shared.ErrNotFound},
		{name: "ErrAlreadyExists matches itself", giveErr: shared.ErrAlreadyExists, wantErr: shared.ErrAlreadyExists},
		{name: "ErrUnauthorized matches itself", giveErr: shared.ErrUnauthorized, wantErr: shared.ErrUnauthorized},
		{name: "ErrForbidden matches itself", giveErr: shared.ErrForbidden, wantErr: shared.ErrForbidden},
		{name: "ErrInvalidInput matches itself", giveErr: shared.ErrInvalidInput, wantErr: shared.ErrInvalidInput},
		{name: "wrapped ErrNotFound matches ErrNotFound", giveErr: errors.Join(shared.ErrNotFound, errors.New("extra")), wantErr: shared.ErrNotFound},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.ErrorIs(t, tt.giveErr, tt.wantErr)
		})
	}
}

func TestPage_ZeroValue(t *testing.T) {
	t.Parallel()

	var p shared.Page
	require.Equal(t, 0, p.Limit)
	require.Empty(t, p.Cursor)
}

func TestPage_Assignment(t *testing.T) {
	t.Parallel()

	p := shared.Page{Limit: 20, Cursor: "abc123"}
	require.Equal(t, 20, p.Limit)
	require.Equal(t, "abc123", p.Cursor)
}

func TestPageResult_NoItems(t *testing.T) {
	t.Parallel()

	result := shared.PageResult[string]{
		Items:      nil,
		NextCursor: "",
		HasMore:    false,
	}
	require.Empty(t, result.Items)
	require.False(t, result.HasMore)
}

func TestPageResult_WithItems(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3}
	result := shared.PageResult[int]{
		Items:      items,
		NextCursor: "cursor-xyz",
		HasMore:    true,
	}
	require.Equal(t, items, result.Items)
	require.Equal(t, "cursor-xyz", result.NextCursor)
	require.True(t, result.HasMore)
}
