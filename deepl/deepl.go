package deepl

import (
	"net/http"
	"net/url"
)

type DeepL interface {
	Translate(Text string, sourceLang string, targetLang string)
}

type DeepLClient struct {
	Endpoint		string	// API endpoint, which differs between the Free and the Pro plans.
	AuthKey			string	// API token, looks like a UUID with ":fx" appended to it.
	LanguagesType	string	// For the "languages" utility call, either "source" or "target".
}

type DeepLResponse struct {
	Translations []Translated
}

type Translated struct {
	DetectedSourceLanguage	string `json:"detected_source_language"`
	Text                 	string `json:"text"`
}

// API call to translate text from sourceLang to targetLang.
func (c *DeepLClient) Translate(text string, sourceLang string, targetLang string) ([]string, error) {
	params := url.Values{}
	params.Add("auth_key", c.AuthKey)
	params.Add("source_lang", sourceLang)
	params.Add("target_lang", targetLang)
	params.Add("text", text)

	var parsed DeepLResponse

	err := c.apiCall(http.MethodPost, params, &parsed)
	if err != nil {
		return []string{}, err
	}
	r := []string{}
	for _, translated := range parsed.Translations {
		r = append(r, translated.Text)
	}
	return r, nil
}

// Returns the base DeepL API endpoint for either the Free or the Pro Plan (if IsPro is true).
func GetEndpoint(isPro bool) string {
	if isPro {
		return "https://api.deepl.com/v2"
	}
	return "https://api-free.deepl.com/v2"
}
