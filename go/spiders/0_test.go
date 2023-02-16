package spiders

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDataKey(t *testing.T) {
	a := getDataKey("{test}")
	require.Equal(t, a, "test")
}
