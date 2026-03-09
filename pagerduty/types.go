package pagerduty

import "time"

// Reference is a minimal PagerDuty object reference.
type Reference struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
	Self    string `json:"self,omitempty"`
	HTMLURL string `json:"html_url,omitempty"`
}

// User represents a PagerDuty user.
type User struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Summary  string `json:"summary"`
	Self     string `json:"self,omitempty"`
	HTMLURL  string `json:"html_url,omitempty"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	TimeZone string `json:"time_zone,omitempty"`
}

// Schedule represents a PagerDuty schedule.
type Schedule struct {
	ID                   string       `json:"id"`
	Type                 string       `json:"type"`
	Summary              string       `json:"summary"`
	Self                 string       `json:"self,omitempty"`
	HTMLURL              string       `json:"html_url,omitempty"`
	Name                 string       `json:"name"`
	TimeZone             string       `json:"time_zone"`
	Description          string       `json:"description,omitempty"`
	Users                []Reference  `json:"users,omitempty"`
	FinalSchedule        *SubSchedule `json:"final_schedule,omitempty"`
	OverridesSubSchedule *SubSchedule `json:"overrides_subschedule,omitempty"`
}

// SubSchedule represents a rendered sub-schedule (final or overrides).
type SubSchedule struct {
	Name                       string               `json:"name"`
	RenderedScheduleEntries    []ScheduleLayerEntry `json:"rendered_schedule_entries"`
	RenderedCoveragePercentage *float64             `json:"rendered_coverage_percentage,omitempty"`
}

// ScheduleLayerEntry represents a single on-call entry in a rendered schedule.
type ScheduleLayerEntry struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	User  Reference `json:"user"`
}

// Override represents a schedule override to create.
type Override struct {
	Start string    `json:"start"`
	End   string    `json:"end"`
	User  Reference `json:"user"`
}

// OverrideResult is the response for a single override creation.
type OverrideResult struct {
	Status   int       `json:"status"`
	Errors   []string  `json:"errors,omitempty"`
	Override *Override `json:"override,omitempty"`
}

// --- API Response wrappers ---

type Pagination struct {
	Offset int  `json:"offset"`
	Limit  int  `json:"limit"`
	More   bool `json:"more"`
	Total  int  `json:"total"`
}

type ListSchedulesResponse struct {
	Pagination
	Schedules []Schedule `json:"schedules"`
}

type GetScheduleResponse struct {
	Schedule Schedule `json:"schedule"`
}

type ListUsersResponse struct {
	Pagination
	Users []User `json:"users"`
}

type CreateOverridesRequest struct {
	Overrides []Override `json:"overrides"`
}
