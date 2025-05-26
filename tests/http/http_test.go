package tests

import (
	"github.com/ljhe/scream/core/socket/http"
	"testing"
)

func TestAcceptor(t *testing.T) {
	http.Server.Start()
}
