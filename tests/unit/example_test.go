package unit

import (
	"testing"

	"github.com/mantit0985/go-template/src/features/example"
)

func TestExampleService(t *testing.T) {
	svc := example.NewService()
	res := svc.DoWork("World")
	expected := "Hello, World! Work completed."

	if res != expected {
		t.Errorf("Expected %s, got %s", expected, res)
	}
}
