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