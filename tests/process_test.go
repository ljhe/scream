package tests

import (
	"context"
	"github.com/ljhe/scream/core/process"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewProcess(t *testing.T) {
	p := process.BuildProcessWithOption(
		process.WithLoader(loader),
	)

	builder := p.System().Loader("mocka").WithID("mocka").WithType("mocka")

	_, err := builder.Register(context.Background())
	assert.Equal(t, err, nil)
}
