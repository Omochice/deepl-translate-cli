package deepl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Generic API call, takes URL parameters and a JSON object to fill,
// validates & parses the response and unmarshals it into the JSON object,
// or throws an error.
// NOTE: Closes the HTTP response that was opened.
func (c *DeepLClient) apiCall(params url.Values, jsonObject any) error {
	resp, err := http.PostForm(c.Endpoint, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return err
	}
	err = parseResponse(resp.Body, jsonObject)
	if err != nil {
		return err
	}
	return nil
}

// Validates the response based on its status code, decoding the returned JSON.
// If the status code is "normal". does nothing (resp remains untouched and open)
func validateResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parsed JSON error data.
		var data map[string]interface{}
		baseErrorText := fmt.Sprintf("Invalid response [%d %s]",
			resp.StatusCode,
			statusText(resp.StatusCode))
		// NOTE: code simplification, we now just use the "standard" error codes from `net/http`.
		// NOTE: on the following code, @Omochice opted for skipping the traditional JSON object struct,
		// going directly for the semi-raw map[string]interface{} reply instead. (gwyneth 20231103)
		e := json.NewDecoder(resp.Body).Decode(&data)
		if e != nil {
			return fmt.Errorf("%s", baseErrorText)
		} else {
			return fmt.Errorf("%s, %s", baseErrorText, data["message"])
		}
	}
	return nil
}

// Generic API response parser, returns whatever the JSON object was.
// The jsonObject to be passed can be anything; returns error from parsing,
// or nil if all's ok.
func parseResponse(resp io.ReadCloser, jsonObject any) error {
	body, err := io.ReadAll(resp)

	if err != nil {
		return fmt.Errorf("%s (occurred while parsing response)", err.Error())
	}
	err = json.Unmarshal(body, &jsonObject)
	if err != nil {
		return fmt.Errorf("%s (occurred while parsing response)", err.Error())
	}
	return nil
}

// Returns an error string based on the HTTP status code, as defined
// by the `net/http` package, plus the additional error(s) from the DeepL API.
// SEE: https://www.deepl.com/docs-api/api-access/error-handling
// NOTE: Replaces former mechanism with a map.
func statusText(statusCode int) string {
	// see if Go knows about this error code:
	respString := http.StatusText(statusCode)	// returns empty string if unknown error code.
	if respString == "" {
		// currently, the DeepL API only adds error code 456, but it may add more in the future...
		switch statusCode {
			case 456:
				respString = "Quota exceeded. The character limit has been reached."
			default:
				respString = "Unknown HTTP error."
		}
	}
	return respString
}
