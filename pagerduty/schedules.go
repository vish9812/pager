package pagerduty

import (
	"fmt"
	"net/url"
	"strconv"
)

// ListSchedules fetches all schedules, handling pagination.
func (c *Client) ListSchedules() ([]Schedule, error) {
	var all []Schedule
	offset := 0
	limit := 100

	for {
		params := url.Values{}
		params.Set("limit", strconv.Itoa(limit))
		params.Set("offset", strconv.Itoa(offset))

		var resp ListSchedulesResponse
		if err := c.get("/schedules", params, &resp); err != nil {
			return nil, fmt.Errorf("listing schedules: %w", err)
		}

		all = append(all, resp.Schedules...)

		if !resp.More {
			break
		}
		offset += limit
	}

	return all, nil
}

// GetSchedule fetches a single schedule with rendered final schedule entries
// for the given time range.
func (c *Client) GetSchedule(id, since, until string) (*Schedule, error) {
	params := url.Values{}
	params.Set("since", since)
	params.Set("until", until)
	params.Set("time_zone", "UTC")

	var resp GetScheduleResponse
	if err := c.get("/schedules/"+id, params, &resp); err != nil {
		return nil, fmt.Errorf("getting schedule %s: %w", id, err)
	}

	return &resp.Schedule, nil
}

// CreateOverrides creates one or more overrides on a schedule.
// Returns the raw JSON response as a slice of OverrideResult.
func (c *Client) CreateOverrides(scheduleID string, overrides []Override) ([]OverrideResult, error) {
	body := CreateOverridesRequest{
		Overrides: overrides,
	}

	var results []OverrideResult
	if err := c.post("/schedules/"+scheduleID+"/overrides", body, &results); err != nil {
		return nil, fmt.Errorf("creating overrides: %w", err)
	}

	return results, nil
}
