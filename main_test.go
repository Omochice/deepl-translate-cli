package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadsettings(t *testing.T) {
	var setting Setting
	var actual Setting
	var err error
	var errorText string

	//
	errorText = "The function should not overload SourceLang / TargetLang if either is not set."
	setting = Setting{
		AuthKey:    "test",
		SourceLang: "EN",
		TargetLang: "JA",
		IsPro:      false,
	}
	actual, err = LoadSettings(setting, false)
	if err != nil {
		t.Fatalf(errorText + "\n%#v", err)
	}
	if setting.AuthKey != actual.AuthKey ||
		setting.SourceLang != actual.SourceLang ||
		setting.TargetLang != actual.TargetLang {
		t.Fatalf(errorText + "\nExpected: %#v\nActual: %#v", setting, actual)
	}

	//
	errorText = "There should occur an error on this function if AuthKey is not set."
	expectedErrorText := "no DeepL token is set; use the environment variable `DEEPL_TOKEN` to set it" // DRY...
	setting = Setting{
		AuthKey:    "",
		SourceLang: "EN",
		TargetLang: "JA",
		IsPro:      false,
	}
	actual, err = LoadSettings(setting, false)
	if err == nil {
		t.Fatalf(errorText+"\nReturned: %#v", actual)
	} else if err.Error() != expectedErrorText {
		t.Fatalf(errorText+"\nExpected: %s\nActual: %s", expectedErrorText, err)
	}

	//
	errorText = "There should occur an error on this function if SourceLang == TargetLang"
	setting = Setting{
		AuthKey:    "test",
		SourceLang: "EN",
		TargetLang: "EN",
		IsPro:      false,
	}
	actual, err = LoadSettings(setting, false)
	if err == nil {
		t.Fatalf(errorText+"\nInput: %#v", setting)
	}
}

// Internal function to test if a file exists or not; this function will *also* be tested below.
func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// This test function first creates a temporary file and checks if Exists correctly identifies it as existing. It then checks a non-existent file path to ensure Exists returns false as expected. You can add this test to your main_test.go file to enhance the reliability of the Exists function.
// As suggested by @coderabbitai (gwyneth 20230413)
func TestExists(t *testing.T) {
	// Test with a file that does exist
	tempFile, err := os.CreateTemp("", "exist_test")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %s", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up after the test

	if !Exists(tempFile.Name()) {
		t.Errorf("Exists should return true for existing file: %s", tempFile.Name())
	}
	defer os.Remove(tempFile.Name()) // Clean up after the test

	// Test with a file that does not exist
	nonExistentFile := filepath.Join(os.TempDir(), "non_existent_file.txt")
	if Exists(nonExistentFile) {
		t.Errorf("Exists should return false for non-existing file: %s", nonExistentFile)
	}
}

func TestInitializeConfigFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "example")
	if err != nil {
		t.Fatalf("Error occurred in os.MkdirTemp: %s", err)
	}
	defer os.RemoveAll(dir)
	p := filepath.Join(dir, "config.json")

	// will success
	if err := InitializeConfigFile(p); err != nil {
		t.Fatalf("There should be no errors thrown by this function\nActual: %s", err)
	}
	if !Exists(p) {
		t.Fatalf("The function should have created the config file: %q", p)
	}
}
