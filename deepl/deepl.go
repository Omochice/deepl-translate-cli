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
	Endpoint string
	AuthKey  string
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

var KnownErrors = map[int]string{
	400: "Bad request. Please check error message and your parameters.",
	403: "Authorization failed. Please supply a valid auth_key parameter.",
	404: "The requested resource could not be found.",
	413: "The request size exceeds the limit.",
	414: "The request URL is too long. You can avoid this error by using a POST request instead of a GET request, and sending the parameters in the HTTP body.",
	429: "Too many requests. Please wait and resend your request.",
	456: "Quota exceeded. The character limit has been reached.",
	503: "Resource currently unavailable. Try again later.",
	529: "Too many requests. Please wait and resend your request.",
} // this from https://www.deepl.com/docs-api/accessing-the-api/error-handling/

func ValidateResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var data map[string]interface{}
		baseErrorText := fmt.Sprintf("Invalid response [%d %s]",
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
		if t, ok := KnownErrors[resp.StatusCode]; ok {
			baseErrorText += fmt.Sprintf(" %s", t)
		}
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
