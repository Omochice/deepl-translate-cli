package deepl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Generic API call, takes method, URL parameters and a JSON object to fill,
// validates & parses the response and unmarshals it into the JSON object,
// or throws an error.
// NOTE: Closes the HTTP response that was opened.
func (c *DeepLClient) apiCall(method string, params url.Values, jsonObject any) error {
	// If we're debugging, show what was printed out:
	if c.Debug > 1 {
		fmt.Fprintf(os.Stderr, "Values being called using %q to API endpoint (%s): %q\n",
			method,
			c.Endpoint,
			params.Encode(),
		)
	}

	// http.PostForm() unfortunately doesn't allow us to set headers, and we need to send the authorization
	// in the headers, not in the body... (gwyneth 20231104)
	client := &http.Client{}
	req, err := http.NewRequest(method, c.Endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "DeepL-Auth-Key " + c.AuthKey)
	resp, err := client.Do(req)
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
// If the status code is "normal", does nothing (`resp` remains untouched and open).
func validateResponse(resp *http.Response) error {
	// http.StatusOK - 200; http.StatusMultipleChoices - 300
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
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
			// Added the response body as suggested by @coderabbitai
			return fmt.Errorf("%s, JSON decoding error was: %s [data received: %v]", baseErrorText, e, resp.Body)
		} else {
			return fmt.Errorf("%s, %s", baseErrorText, data["message"])
		}
	}
	return nil
}

// Generic API response parser, returns whatever the JSON object was.
// The `jsonObject` to be passed can be anything; returns error from parsing,
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
