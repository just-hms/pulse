package engine_test

import (
	"testing"

	"github.com/just-hms/pulse/pkg/engine"
	"github.com/stretchr/testify/require"
)

func TestParseMetric(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	m, err := engine.ParseMetric("BM25")
	req.NoError(err)
	req.Equal(m, engine.BM25)
}
