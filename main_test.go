package main

import (
	"io/ioutil"
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
	errorText = "The function should not overload SourceLang / TargetLang if eigher is not set."
	setting = Setting{
		AuthKey:    "test",
		SourceLang: "EN",
		TargetLang: "JA",
		IsPro:      false,
	}
	actual, err = LoadSettings(setting, false)
	if err != nil {
		t.Fatalf(errorText+"\n%#v", err)
	}
	if setting.AuthKey != actual.AuthKey ||
		setting.SourceLang != actual.SourceLang ||
		setting.TargetLang != actual.TargetLang {
		t.Fatalf(errorText+"\nExpected: %#v\nActual: %#v", setting, actual)
	}

	//
	errorText = "The function should occur error if AuthKey is not set."
	expectedErrorText := "No deepl token is set." // DRY...
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
		t.Fatalf(errorText+"\nExpected: %s\nActual: %s", expectedErrorText, err.Error())
	}

	//
	errorText = "The function should occur error if SourceLang == TargetLang"
	setting = Setting{
		AuthKey:    "test",
		SourceLang: "EN",
		TargetLang: "EN",
		IsPro:      false,
	}
	actual, err = LoadSettings(setting, false)
	if err == nil {
		t.Fatalf(errorText+"\nInputed: %#v", setting)
	}
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func TestInitializeConfigFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatalf("Error occurred in iotuil.Tempdir")
	}
	defer os.RemoveAll(dir)
	p := filepath.Join(dir, "config.json")

	// will success
	if err := InitializeConfigFile(p); err != nil {
		t.Fatalf("The function should not occur error\nActual: %s", err.Error())
	}
	if !Exists(p) {
		t.Fatalf("The function should make configfile(%s)", p)
	}
}
