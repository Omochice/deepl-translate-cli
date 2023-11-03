// This file handles non-translation calls, namely for usage, languages supported, and so forth.
package deepl

import (
	//	"encoding/json"
	"fmt"
	"net/url"
)

type DeepLUsageResponse struct {
	CharacterCount		int	`json:"character_count,omitempty"`// Characters translated so far in the current billing period.
	CharacterLimit		int	`json:"character_limit,omitempty"`// Current maximum number of characters that can be translated per billing period.
	DocumentLimit		int	`json:"document_limit,omitempty"` // Documents translated so far in the current billing period.
	DocumentCount		int	`json:"document_count,omitempty"` // Current maximum number of documents that can be translated per billing period.
	TeamDocumentLimit	int	`json:"team_document_limit,omitempty"` // Documents translated by all users in the team so far in the current billing period.
	TeamDocumentCount	int	`json:"team_document_count,omitempty"` // Current maximum number of documents that can be translated by the team per billing period.
}

// Check Usage and Limits —
// Retrieve usage information within the current billing period together with the corresponding account limits.
func (c *DeepLClient) Usage() (string, error) {
	params := url.Values{}
	params.Add("auth_key", c.AuthKey)

	var resp DeepLUsageResponse

	err := c.apiCall(params, &resp)
	if err != nil {
		return "", err
	}

 	return fmt.Sprintf(
		"Character Count: %d; Character Limit: %d; Document Limit: %d; Document Count: %d; Team Document Limit: %d; Team Document Count: %d.",
		resp.CharacterCount,
		resp.CharacterLimit,
		resp.DocumentLimit,
		resp.DocumentCount,
		resp.TeamDocumentLimit,
		resp.TeamDocumentCount),
	nil
}

// The /languages API call returns an array of language/name pairs and a flag
// indicating if this language has support for formal/informal differences.
type DeepLLanguagesResponse struct {
	Language			string	`json:"language"`
	Name				string	`json:"name"`
	SupportsFormality	bool	`json:"supports_formality"`
}

// Retrieve Supported Languages —
// Retrieve the list of languages that are currently supported for translation, either as source or target language, respectively.
func (c *DeepLClient) Languages() (string, error) {
	params := url.Values{}
	params.Add("auth_key", c.AuthKey)
	// note: translationType was already parsed
	if len(c.LanguagesType) > 0 {
		params.Add("type", c.LanguagesType)
	}

	var langs []DeepLLanguagesResponse

	err := c.apiCall(params, &langs)
	if err != nil {
		return "", err
	}

	var r string
	for _, lang := range langs {
		r += lang.Language + ": " + lang.Name
		if lang.SupportsFormality {
			r += " (+ formality)"
		}
		r += "\n"
	}
	return r, nil
}