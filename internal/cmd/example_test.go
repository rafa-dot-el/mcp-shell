package cmd

import (
	"testing"
)

func TestExampleFunction(t *testing.T) {
	// Simple test to demonstrate PR testing feedback
	result := exampleFunction("test")
	expected := "Hello, test!"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestExampleFunctionEmpty(t *testing.T) {
	// Another test to increase coverage
	result := exampleFunction("")
	expected := "Hello, !"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// Example function to test
func exampleFunction(name string) string {
	return "Hello, " + name + "!"
}
