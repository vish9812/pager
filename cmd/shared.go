package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"

	"pager/cache"
	"pager/pagerduty"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	tableHeader  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).Padding(0, 2)
	tableCell    = lipgloss.NewStyle().Padding(0, 2)
)

// --- Data loading (with cache) ---

// loadSchedulesAndUsers returns schedules and users, using cache if available.
func loadSchedulesAndUsers() ([]pagerduty.Schedule, []pagerduty.User, error) {
	cached := cache.Load()
	if cached != nil {
		age := time.Since(cached.FetchedAt)
		fmt.Println(dimStyle.Render(fmt.Sprintf("  Using cached data (fetched %s ago). Run 'pager cache clear' to refresh.", formatAge(age))))
		return cached.Schedules, cached.Users, nil
	}

	var schedules []pagerduty.Schedule
	var users []pagerduty.User
	var schedErr, userErr error

	err := spinner.New().
		Title("Fetching schedules and users...").
		Action(func() {
			schedules, schedErr = pdClient.ListSchedules()
			if schedErr == nil {
				users, userErr = pdClient.ListUsers()
			}
		}).
		Run()
	if err != nil {
		return nil, nil, fmt.Errorf("spinner error: %w", err)
	}
	if schedErr != nil {
		return nil, nil, fmt.Errorf("fetching schedules: %w", schedErr)
	}
	if userErr != nil {
		return nil, nil, fmt.Errorf("fetching users: %w", userErr)
	}

	// Save to cache (best-effort, don't fail the command)
	_ = cache.Save(schedules, users)

	return schedules, users, nil
}

// --- Interactive prompts ---

// selectSchedule shows a filterable schedule picker. Defaults to the last selected schedule.
func selectSchedule(schedules []pagerduty.Schedule) (string, error) {
	if len(schedules) == 0 {
		return "", fmt.Errorf("no schedules found in your PagerDuty account")
	}

	options := make([]huh.Option[string], len(schedules))
	for i, s := range schedules {
		label := s.Name
		if s.Description != "" {
			label += dimStyle.Render(" - " + s.Description)
		}
		options[i] = huh.NewOption(label, s.ID)
	}

	scheduleID := cache.LoadLastScheduleID()
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a schedule").
				Options(options...).
				Filtering(true).
				Value(&scheduleID),
		),
	).Run()
	if err != nil {
		return "", err
	}

	// Save selection for next run (best-effort)
	_ = cache.SaveLastScheduleID(scheduleID)

	return scheduleID, nil
}

// selectUser shows a filterable user picker with the given title.
func selectUser(title string, users []pagerduty.User) (string, error) {
	if len(users) == 0 {
		return "", fmt.Errorf("no users found in your PagerDuty account")
	}

	options := make([]huh.Option[string], len(users))
	for i, u := range users {
		label := u.Name + dimStyle.Render(" <"+u.Email+">")
		options[i] = huh.NewOption(label, u.ID)
	}

	var userID string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(options...).
				Filtering(true).
				Height(15).
				Value(&userID),
		),
	).Run()
	if err != nil {
		return "", err
	}

	return userID, nil
}

// enterDateRange prompts for start and end dates in two sequential forms.
// After entering the start date and pressing tab/enter, the end date field
// is pre-populated with start + 7 days.
func enterDateRange() (string, string, error) {
	now := time.Now()
	defaultSince := now.Format("2006-01-02")

	// --- Form 1: Start date ---
	sinceStr := defaultSince
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Start date").
				Description("Format: YYYY-MM-DD or YYYY-MM-DDTHH:MM").
				Value(&sinceStr).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("start date is required")
					}
					_, err := parseDateTime(s)
					return err
				}),
		),
	).Run()
	if err != nil {
		return "", "", err
	}

	sinceTime, err := parseDateTime(sinceStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid start date: %w", err)
	}

	// Compute default end date: start + 7 days
	defaultUntil := sinceTime.AddDate(0, 0, 7).Format("2006-01-02")

	// --- Form 2: End date (with computed default) ---
	untilStr := defaultUntil
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("End date").
				Description("Format: YYYY-MM-DD or YYYY-MM-DDTHH:MM").
				Value(&untilStr).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("end date is required")
					}
					_, err := parseDateTime(s)
					return err
				}),
		),
	).Run()
	if err != nil {
		return "", "", err
	}

	untilTime, err := parseDateTime(untilStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid end date: %w", err)
	}

	if !untilTime.After(sinceTime) {
		return "", "", fmt.Errorf("end date must be after start date")
	}

	return sinceTime.UTC().Format(time.RFC3339), untilTime.UTC().Format(time.RFC3339), nil
}

// fetchScheduleEntries fetches a schedule with rendered entries for a date range.
func fetchScheduleEntries(scheduleID, since, until string) (*pagerduty.Schedule, error) {
	var schedule *pagerduty.Schedule
	var fetchErr error

	err := spinner.New().
		Title("Fetching schedule entries...").
		Action(func() {
			schedule, fetchErr = pdClient.GetSchedule(scheduleID, since, until)
		}).
		Run()
	if err != nil {
		return nil, fmt.Errorf("spinner error: %w", err)
	}
	if fetchErr != nil {
		return nil, fmt.Errorf("fetching schedule: %w", fetchErr)
	}

	return schedule, nil
}

// --- Display helpers ---

// printEntries prints on-call entries for a single known user (no user column).
func printEntries(entries []pagerduty.ScheduleLayerEntry) {
	if len(entries) == 0 {
		fmt.Println(dimStyle.Render("  No entries."))
		fmt.Println()
		return
	}

	fmt.Printf("  %s%s%s\n",
		tableHeader.Render("Start"),
		tableHeader.Render("End"),
		tableHeader.Render("Duration"),
	)
	fmt.Printf("  %s%s%s\n",
		tableHeader.Render(strings.Repeat("─", 21)),
		tableHeader.Render(strings.Repeat("─", 21)),
		tableHeader.Render(strings.Repeat("─", 10)),
	)

	localZone := time.Now().Location()
	for _, e := range entries {
		start := e.Start.In(localZone)
		end := e.End.In(localZone)
		duration := end.Sub(start)

		fmt.Printf("  %s%s%s\n",
			tableCell.Render(start.Format("2006-01-02 15:04 MST")),
			tableCell.Render(end.Format("2006-01-02 15:04 MST")),
			tableCell.Render(formatDuration(duration)),
		)
	}
	fmt.Println()
}

// printAllEntries prints on-call entries with a user column (for oncall command).
func printAllEntries(entries []pagerduty.ScheduleLayerEntry) {
	if len(entries) == 0 {
		fmt.Println(dimStyle.Render("  No entries found in this date range."))
		fmt.Println()
		return
	}

	fmt.Printf("  %s%s%s%s\n",
		tableHeader.Render(padRight("User", 24)),
		tableHeader.Render("Start"),
		tableHeader.Render("End"),
		tableHeader.Render("Duration"),
	)
	fmt.Printf("  %s%s%s%s\n",
		tableHeader.Render(strings.Repeat("─", 24)),
		tableHeader.Render(strings.Repeat("─", 21)),
		tableHeader.Render(strings.Repeat("─", 21)),
		tableHeader.Render(strings.Repeat("─", 10)),
	)

	localZone := time.Now().Location()
	for _, e := range entries {
		start := e.Start.In(localZone)
		end := e.End.In(localZone)
		duration := end.Sub(start)

		fmt.Printf("  %s%s%s%s\n",
			tableCell.Render(padRight(e.User.Summary, 24)),
			tableCell.Render(start.Format("2006-01-02 15:04 MST")),
			tableCell.Render(end.Format("2006-01-02 15:04 MST")),
			tableCell.Render(formatDuration(duration)),
		)
	}
	fmt.Println()
}

// --- Filtering ---

func filterEntriesForUser(schedule *pagerduty.Schedule, userID string) []pagerduty.ScheduleLayerEntry {
	if schedule.FinalSchedule == nil {
		return nil
	}

	var filtered []pagerduty.ScheduleLayerEntry
	for _, entry := range schedule.FinalSchedule.RenderedScheduleEntries {
		if entry.User.ID == userID {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// --- Helpers ---

func findUser(users []pagerduty.User, id string) *pagerduty.User {
	for _, u := range users {
		if u.ID == id {
			return &u
		}
	}
	return &pagerduty.User{ID: id, Name: "Unknown"}
}

func parseDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	if t, err := time.ParseInLocation("2006-01-02T15:04", s, time.Now().Location()); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Now().Location()); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Now().Location()); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date format %q, expected YYYY-MM-DD or YYYY-MM-DDTHH:MM", s)
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours >= 24 {
		days := hours / 24
		remainingHours := hours % 24
		if remainingHours == 0 && minutes == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd %dh", days, remainingHours)
	}
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func formatAge(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}
