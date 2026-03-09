package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"pager/pagerduty"
)

func init() {
	rootCmd.AddCommand(overrideCmd)
}

var overrideCmd = &cobra.Command{
	Use:   "override",
	Short: "Override a user's on-call schedule",
	Long:  "Interactively override a user's on-call shifts with another user for a given date range.",
	RunE:  runOverride,
}

func runOverride(cmd *cobra.Command, args []string) error {
	// Load schedules and users (from cache or API)
	schedules, users, err := loadSchedulesAndUsers()
	if err != nil {
		return err
	}

	// Step 1: Select schedule
	scheduleID, err := selectSchedule(schedules)
	if err != nil {
		return err
	}

	// Step 2: Select user to replace
	targetUserID, err := selectUser("Who do you want to override (replace)?", users)
	if err != nil {
		return err
	}

	// Step 3: Enter date range
	since, until, err := enterDateRange()
	if err != nil {
		return err
	}

	// Step 4: Fetch schedule entries for date range
	schedule, err := fetchScheduleEntries(scheduleID, since, until)
	if err != nil {
		return err
	}

	// Step 5: Filter entries for the target user
	entries := filterEntriesForUser(schedule, targetUserID)
	targetUser := findUser(users, targetUserID)

	if len(entries) == 0 {
		fmt.Println()
		fmt.Println(errorStyle.Render("  User " + targetUser.Name + " is not on-call in the specified date range."))
		fmt.Println()
		return nil
	}

	// Step 6: Display current on-call entries
	fmt.Println()
	fmt.Println(headerStyle.Render("  Current on-call entries for " + targetUser.Name + ":"))
	printEntries(entries)

	// Step 7: Select replacement user
	fmt.Println(headerStyle.Render("  Select who should take over these shifts:"))
	replacementUserID, err := selectUser("Override with", users)
	if err != nil {
		return err
	}

	// Step 8: Confirm
	replacementUser := findUser(users, replacementUserID)
	confirmed, err := confirmOverride(targetUser, replacementUser, entries)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println(dimStyle.Render("  Override cancelled."))
		return nil
	}

	// Step 9: Create overrides
	overrides := buildOverrides(entries, replacementUserID)
	results, err := createScheduleOverrides(scheduleID, overrides)
	if err != nil {
		return err
	}

	// Check results for errors
	hasErrors := false
	for _, r := range results {
		if r.Status < 200 || r.Status >= 300 {
			hasErrors = true
			fmt.Println(errorStyle.Render(fmt.Sprintf("  Override failed: %s", strings.Join(r.Errors, ", "))))
		}
	}
	if hasErrors {
		return fmt.Errorf("some overrides failed, see above")
	}

	fmt.Println()
	fmt.Println(successStyle.Render("  Overrides created successfully!"))
	fmt.Println()

	// Step 10: Re-fetch and display updated schedule
	updatedSchedule, err := fetchScheduleEntries(scheduleID, since, until)
	if err != nil {
		return fmt.Errorf("fetching updated schedule: %w", err)
	}

	updatedEntries := filterEntriesForUser(updatedSchedule, replacementUserID)
	fmt.Println(headerStyle.Render("  Updated on-call entries for " + replacementUser.Name + ":"))
	printEntries(updatedEntries)

	return nil
}

// --- Override-specific helpers ---

func confirmOverride(target, replacement *pagerduty.User, entries []pagerduty.ScheduleLayerEntry) (bool, error) {
	summary := fmt.Sprintf(
		"Replace %s with %s for %d shift(s)?",
		target.Name, replacement.Name, len(entries),
	)

	var confirmed bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(summary).
				Affirmative("Yes, override").
				Negative("Cancel").
				Value(&confirmed),
		),
	).Run()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

func buildOverrides(entries []pagerduty.ScheduleLayerEntry, replacementUserID string) []pagerduty.Override {
	overrides := make([]pagerduty.Override, len(entries))
	for i, e := range entries {
		overrides[i] = pagerduty.Override{
			Start: e.Start.UTC().Format(time.RFC3339),
			End:   e.End.UTC().Format(time.RFC3339),
			User: pagerduty.Reference{
				ID:   replacementUserID,
				Type: "user_reference",
			},
		}
	}
	return overrides
}

func createScheduleOverrides(scheduleID string, overrides []pagerduty.Override) ([]pagerduty.OverrideResult, error) {
	var results []pagerduty.OverrideResult
	var createErr error

	err := spinner.New().
		Title("Creating overrides...").
		Action(func() {
			results, createErr = pdClient.CreateOverrides(scheduleID, overrides)
		}).
		Run()
	if err != nil {
		return nil, fmt.Errorf("spinner error: %w", err)
	}
	if createErr != nil {
		return nil, fmt.Errorf("creating overrides: %w", createErr)
	}

	return results, nil
}
