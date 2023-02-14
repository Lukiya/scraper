package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadRules(t *testing.T) {
	rules, err := ReadRulesFromFile("../rules_sample.json")
	require.NoError(t, err)
	require.NotEmpty(t, rules)
}
