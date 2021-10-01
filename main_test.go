package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
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

func TestValidateResponse(t *testing.T) {
	var errorText string
	testErrorMes := map[string]string{
		"message": "test message",
	}
	testBody, err := json.Marshal(testErrorMes)
	if err != nil {
		t.Fatalf("Error within json.Marshal")
	}
	testResponse := http.Response{
		Status:     "200 test",
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(testBody)),
	}

	for c := 100; c < 512; c++ {
		if http.StatusText(c) == "" {
			// unused stasus code
			continue
		}
		testResponse.Status = fmt.Sprintf("%d test", c)
		testResponse.StatusCode = c
		err := ValidateResponse(&testResponse)
		if c >= 200 && c < 300 {
			errorText = "If status code is between 200 and 299, no error should be returned"
			if err != nil {
				t.Fatalf("%s\nActual: %s", errorText, err.Error())
			}
			continue
		}
		if err == nil {
			t.Fatalf("If Status code is not 200 <= c < 300, should occur error\nStatus code: %d, Response: %v",
				c, testResponse)
		} else {
			if !strings.Contains(err.Error(), http.StatusText(c)) {
				errorText = fmt.Sprintf("Error text should include Status Code(%s)\nActual: %s",
					http.StatusText(c), err.Error())
			}
			if statusText, ok := KnownErrors[c]; ok { // errored
				if !strings.Contains(err.Error(), statusText) {
					errorText = fmt.Sprintf("If stasus code is knownded, the text should include it's error text\nExpected: %s\nActual: %s",
						statusText, err.Error())
					t.Fatalf("%s", errorText)
				}
			}
		}
	}
	// test when body is valid/invalid as json
	invalidResp := http.Response{
		Status:     "444 not exists error",
		StatusCode: 444,
		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("test"))),
	}
	err = ValidateResponse(&invalidResp)
	if err == nil {
		t.Fatalf("If status code is invalid, should occur error\nActual: %d", invalidResp.StatusCode)
	} else if !strings.HasSuffix(err.Error(), "]") {
		t.Fatalf("If body is invalid as json, error is formated as %s",
			"`Invalid response [statuscode statustext]`")
	}

	expectedMessage := "This is test"
	validResp := http.Response{
		Status:     "444 not exists error",
		StatusCode: 444,
		Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(`{"message": "%s"}`, expectedMessage)))),
	}
	err = ValidateResponse(&validResp)
	if err == nil {
		t.Fatalf("If is status code invalid, should occur error\nActual: %d", validResp.StatusCode)
	} else if !strings.HasSuffix(err.Error(), expectedMessage) {
		t.Fatalf("If body is valid as json, error suffix should have `%s`\nActual: %s",
			expectedMessage, err.Error())
	}
}

func TestParseResponse(t *testing.T) {
	baseResponse := http.Response{
		Status:     "200 test",
		StatusCode: 200,
	}
	{
		input := map[string][]map[string]string{
			"Translations": make([]map[string]string, 3),
		}
		sampleRes := map[string]string{
			"detected_source_language": "test",
			"text":                     "test text",
		}
		input["Translations"][0] = sampleRes
		input["Translations"][1] = sampleRes
		input["Translations"][2] = sampleRes
		b, err := json.Marshal(input)
		if err != nil {
			t.Fatal("Error within json.Marshal")
		}
		baseResponse.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		res, err := ParseResponse(&baseResponse)
		if err != nil {
			t.Fatalf("If input is valid, should not occur error\n%s", err.Error())
		}

		if len(res.Translations) != len(input["Translations"]) {
			t.Fatalf("Length of result.Translations should be equal to input.Translations\nExpected: %d\nActual: %d",
				len(res.Translations), len(input["Translations"]))
		}
	}
	{
		input := map[string][]map[string]string{
			"Translations": make([]map[string]string, 1),
		}
		sampleRes := map[string]string{
			"detected_source_language": "test",
			"text":                     "test text",
			"this will be ignored":     "test",
		}
		input["Translations"][0] = sampleRes
		b, err := json.Marshal(input)
		if err != nil {
			t.Fatal("Error within json.Marshal")
		}
		baseResponse.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		res, err := ParseResponse(&baseResponse)
		if err != nil {
			t.Fatalf("If input is valid, should not occur error\n%s", err.Error())
		}
		if len(res.Translations) != len(input["Translations"]) {
			t.Fatalf("Length of result.Translations should be equal to input.Translations\nExpected: %d\nActual: %d",
				len(res.Translations), len(input["Translations"]))
		}
		resType := reflect.ValueOf(res.Translations[0]).Type()
		expectedNumOfField := 2
		if resType.NumField() != expectedNumOfField {
			t.Fatalf("Length of Translated's filed should be equal %d\nActual: %d", expectedNumOfField, resType.NumField())
		}
	}
}
