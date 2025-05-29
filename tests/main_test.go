package tests

import (
	"fmt"
	"os"
	"testing"
)

func init() {
	fmt.Println("main test init")
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
