// This file handles non-translation calls, namely for usage, languages supported, and so forth.
package deepl

import (
	"net/url"
)

// Check Usage and Limits —
// Retrieve usage information within the current billing period together with the corresponding account limits.
func (c *DeepLClient) Usage() (string, error) {
	params := url.Values{}
	params.Add("auth_key", c.AuthKey)

	parsed, err := c.apiCall(params)
	if err != nil {
		return "", err
	}
	return parsed.(string), nil
}