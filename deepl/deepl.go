package deepl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type DeepL interface {
	Translate(Text string, sourceLang string, targetLang string)
}

type DeepLClient struct {
	Endpoint string		// API endpoint, which differs between the Free and the Pro plans.
	AuthKey  string		// API token, looks like a UUID with ":fx" appended to it.
}

type DeepLResponse struct {
	Translations []Translated
}

type Translated struct {
	DetectedSourceLaguage string `json:"detected_source_language"`
	Text                  string `json:"text"`
}

func (c *DeepLClient) Translate(text string, sourceLang string, targetLang string) ([]string, error) {
	params := url.Values{}
	params.Add("auth_key", c.AuthKey)
	params.Add("source_lang", sourceLang)
	params.Add("target_lang", targetLang)
	params.Add("text", text)
	resp, _ := http.PostForm(c.Endpoint, params)

	if err := ValidateResponse(resp); err != nil {
		return []string{}, err
	}
	parsed, err := ParseResponse(resp)
	if err != nil {
		return []string{}, err
	}
	r := []string{}
	for _, translated := range parsed.Translations {
		r = append(r, translated.Text)
	}
	return r, nil
}

// Returns an error string based on the HTTP status code, as defined
// by the `net/http` package, plus the additional error(s) from the DeepL API.
// SEE: https://www.deepl.com/docs-api/api-access/error-handling
// NOTE: Replaces former mechanism with a map.
func StatusText(statusCode int) string {
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

// Validates the response based on its status code, decoding the returned JSON.
func ValidateResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parsed JSON error data.
		var data map[string]interface{}
		baseErrorText := fmt.Sprintf("Invalid response [%d %s]",
			resp.StatusCode,
			StatusText(resp.StatusCode))
		// NOTE: code simplification, we now just use the "standard" error codes from `net/htp`.
		e := json.NewDecoder(resp.Body).Decode(&data)
		if e != nil {
			return fmt.Errorf("%s", baseErrorText)
		} else {
			return fmt.Errorf("%s, %s", baseErrorText, data["message"])
		}
	}
	return nil
}

func ParseResponse(resp *http.Response) (DeepLResponse, error) {
	var responseJson DeepLResponse
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		err := fmt.Errorf("%s (occurred while parsing response)", err.Error())
		return responseJson, err
	}
	err = json.Unmarshal(body, &responseJson)
	if err != nil {
		err := fmt.Errorf("%s (occurred while parsing response)", err.Error())
		return responseJson, err
	}
	return responseJson, err
}

func GetEndpoint(IsPro bool) string {
	if IsPro {
		return "https://api.deepl.com/v2/translate"
	}
	return "https://api-free.deepl.com/v2/translate"
}
