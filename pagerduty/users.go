package pagerduty

import (
	"fmt"
	"net/url"
	"strconv"
)

// ListUsers fetches all users in the account, handling pagination.
func (c *Client) ListUsers() ([]User, error) {
	var all []User
	offset := 0
	limit := 100

	for {
		params := url.Values{}
		params.Set("limit", strconv.Itoa(limit))
		params.Set("offset", strconv.Itoa(offset))

		var resp ListUsersResponse
		if err := c.get("/users", params, &resp); err != nil {
			return nil, fmt.Errorf("listing users: %w", err)
		}

		all = append(all, resp.Users...)

		if !resp.More {
			break
		}
		offset += limit
	}

	return all, nil
}
