package deepl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

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
		Status:     fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)), // 200 OK
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(testBody)),
	}

	for c := 100; c < 512; c++ {
		if http.StatusText(c) == "" {
			// unused status code
			continue
		}
		testResponse.Status = fmt.Sprintf("%d test", c)
		testResponse.StatusCode = c
		err := validateResponse(&testResponse)
		if c >= 200 && c < 300 {
			errorText = "If status code is between 200 and 299, no error should be returned"
			if err != nil {
				t.Fatalf("%s\nActual: %s", errorText, err.Error())
			}
			continue
		}
		if err == nil {
			t.Fatalf("If status code is not 200 <= c < 300, an error should occur\nStatus code: %d, Response: %v",
				c, testResponse)
		} else {
			if !strings.Contains(err.Error(), http.StatusText(c)) {
				errorText = fmt.Sprintf("Error text should include Status Code(%s)\nActual: %s",
					http.StatusText(c), err.Error())
					t.Fatalf("%s", errorText)	// was this missing?
			}
			statusText := fmt.Sprintf("%d", c)	// NOTE: we want to make sure this is a string, not a (single) rune (gwyneth 20231101)
			if !strings.Contains(err.Error(), statusText) {
				errorText = fmt.Sprintf("If status code is known, the text should include its error text\nExpected: %s\nActual: %s",
					statusText, err.Error())
				t.Fatalf("%s", errorText)
			}
		}
	}
	// test when body is valid/invalid as json
	invalidResp := http.Response{
		Status:     "444 not exists error",
		StatusCode: 444,
		Body:       io.NopCloser(bytes.NewBuffer([]byte("test"))),
	}
	err = validateResponse(&invalidResp)
	if err == nil {
		t.Fatalf("If status code is invalid, an error should occur\nActual: %d", invalidResp.StatusCode)
	} else if !strings.HasSuffix(err.Error(), "]") {
		t.Fatalf("If body is invalid as JSON, error is formatted as %s",
			"`Invalid response [statuscode statustext]`")
	}

	expectedMessage := "This is test"
	validResp := http.Response{
		Status:     "444 not exists error",
		StatusCode: 444,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(`{"message": "%s"}`, expectedMessage)))),
	}
	err = validateResponse(&validResp)
	if err == nil {
		t.Fatalf("If the status code is invalid, an error should occur\nActual: %d", validResp.StatusCode)
	} else if !strings.HasSuffix(err.Error(), expectedMessage) {
		t.Fatalf("If the body is valid JSON, the error suffix should have `%s`\nActual: %s",
			expectedMessage, err.Error())
	}
}

func TestParseTranslationResponse(t *testing.T) {
	baseResponse := http.Response{
		Status:     fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)), //  200 OK
		StatusCode: http.StatusOK,	// 200
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
		baseResponse.Body = io.NopCloser(bytes.NewBuffer(b))
		var trans DeepLResponse
		err = parseResponse(baseResponse.Body, &trans)
		if err != nil {
			t.Fatalf("If the input is valid, no errors should occur\n%s", err.Error())
		}
		if len(trans.Translations) != len(input["Translations"]) {
			t.Fatalf("Length of result.Translations should be equal to input.Translations\nExpected: %d\nActual: %d",
				len(trans.Translations), len(input["Translations"]))
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
		baseResponse.Body = io.NopCloser(bytes.NewBuffer(b))
		var trans DeepLResponse
		err = parseResponse(baseResponse.Body, &trans)
		if err != nil {
			t.Fatalf("If the input is valid, no error should occur\n%s", err.Error())
		}
		if len(trans.Translations) != len(input["Translations"]) {
			t.Fatalf("Length of result.Translations should be equal to input.Translations\nExpected: %d\nActual: %d",
				len(trans.Translations), len(input["Translations"]))
		}
		resType := reflect.ValueOf(trans.Translations[0]).Type()
		expectedNumOfField := 2
		if resType.NumField() != expectedNumOfField {
			t.Fatalf("Length of translated field should be equal to %d\nActual: %d", expectedNumOfField, resType.NumField())
		}
	}
}

// added by @coderabbitai
func TestValidateResponseEdgeCases(t *testing.T) {
	// Test case: Empty response body
	emptyResp := http.Response{
		Status:     fmt.Sprintf("%d %s", http.StatusNoContent, http.StatusText(http.StatusNoContent)), // 204 No Content
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
	}
	err := validateResponse(&emptyResp)
	if err != nil {
		t.Errorf("Expected no error for empty response body, got: %v", err)
	}

	// Test case: Non-JSON content in response body
	// Note: this test has to be done with status code below 200 or over 300! (gwyneth 20240413)
	nonJSONResp := http.Response{
		// Status:     "200 OK",
		// StatusCode: 200,
		Status:     fmt.Sprintf("%d %s", http.StatusNotFound, http.StatusText(http.StatusNotFound)),	// 404 Not Found
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBuffer([]byte("Not a JSON response"))),
	}
	err = validateResponse(&nonJSONResp)
	if err == nil {
		t.Errorf("Expected error for non-JSON response body; got no error instead")
	} else if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("Unexpected error for non-JSON response body: %v", err)
	}
	// Test case: JSON response with unexpected fields
	unexpectedJSONResp := http.Response{
		Status:     fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)), // 200 OK
		StatusCode: http.StatusOK,	// 200
		Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"unexpected": "field"}`))),
	}
	err = validateResponse(&unexpectedJSONResp)
	if err != nil {
		t.Errorf("Expected no error for JSON with unexpected fields, got: %v", err)
	}

	// Test case: Network error simulation (this would typically be handled outside this function, but included for completeness).
	// This case would need to mock network failure, which is outside the scope of validateResponse function as it does not handle netwrk operations directly.
}

// A preliminary test for the Translate() function. I'm no good at this, so... baby steps.
// (gwyneth 20240412)
func TestTranslate(t *testing.T) {
	var deeplTestClient DeepLClient

	// Translating an empty string should give an error.
	translateds, err := deeplTestClient.Translate("")
	if err == nil {
		t.Errorf("Expected an error from attempt to translate empty string")
	} else {
		t.Logf("✅🆗 Attempt to translate an empty string returned expected error %v", err)
	}
	if translateds != nil {
		t.Errorf("The translation ought to have been empty, but got the following instead: %v", translateds)
	}
}