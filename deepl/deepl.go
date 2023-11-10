package deepl

import (
	"net/http"
	"net/url"
	"strconv"
)

type DeepL interface {
	Translate(Text string)
}

type DeepLClient struct {
	Endpoint			string	`json:"endpoint"`				// API endpoint, which differs between the Free and the Pro plans.
	AuthKey				string	`json:"authkey"`				// API token, looks like a UUID with ":fx". appended to it.
	SourceLang 			string	`json:"source_lang"`
	TargetLang 			string	`json:"target_lang"`
	LanguagesType		string	`json:"type"`					// For the "languages" utility call, either "source" or "target".
	IsPro      			bool	`json:"-"`
	TagHandling			string	`json:"tag_handling"`			// "xml", "html".
	SplitSentences		string	`json:"split_sentences"`		// "0", "1", "norewrite".
	PreserveFormatting	string	`json:"preserve_formatting"`	// "0", "1".
	OutlineDetection	int		`json:"outline_detection"`		// Integer; 0 is default.
	NonSplittingTags	string	`json:"non_splitting_tags"`		// List of comma-separated XML tags.
	SplittingTags		string	`json:"splitting_tags"`			// List of comma-separated XML tags.
	IgnoreTags			string	`json:"ignore_tags"`			// List of comma-separated XML tags.
	Debug				int		`json:"debug"`					// Debug/verbosity level, 0 is no debugging
}

type DeepLResponse struct {
	Translations []Translated
}

type Translated struct {
	DetectedSourceLanguage	string `json:"detected_source_language"`
	Text                 	string `json:"text"`
}

// API call to translate text from sourceLang to targetLang.
func (c *DeepLClient) Translate(text string) ([]string, error) {
	// TODO(gwyneth): Make the call with JSON, it probably makes much more sense that way.
	params := url.Values{}
	params.Add("auth_key",				c.AuthKey)
	params.Add("source_lang",			c.SourceLang)
	params.Add("target_lang",			c.TargetLang)
	params.Add("type",					c.LanguagesType)
	params.Add("tag_handling",			c.TagHandling)
	params.Add("split_sentences",		c.SplitSentences)
	params.Add("preserve_formatting",	c.PreserveFormatting)
	params.Add("outline_detection",		strconv.Itoa(c.OutlineDetection))
	params.Add("non_splitting_tags",	c.NonSplittingTags)
	params.Add("splitting_tags",		c.SplittingTags)
	params.Add("ignore_tags",			c.IgnoreTags)
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
