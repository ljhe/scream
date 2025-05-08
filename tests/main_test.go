package tests

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Println("main testing ...")
	code := m.Run()
	os.Exit(code)
}
