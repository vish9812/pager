package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(oncallCmd)
}

var oncallCmd = &cobra.Command{
	Use:   "oncall",
	Short: "Check who is on-call for a schedule",
	Long:  "View all on-call assignments for a given schedule and date range.",
	RunE:  runOncall,
}

func runOncall(cmd *cobra.Command, args []string) error {
	// Load schedules and users (from cache or API)
	schedules, _, err := loadSchedulesAndUsers()
	if err != nil {
		return err
	}

	// Step 1: Select schedule
	scheduleID, err := selectSchedule(schedules)
	if err != nil {
		return err
	}

	// Step 2: Enter date range
	since, until, err := enterDateRange()
	if err != nil {
		return err
	}

	// Step 3: Fetch schedule entries
	schedule, err := fetchScheduleEntries(scheduleID, since, until)
	if err != nil {
		return err
	}

	// Step 4: Display all on-call entries
	if schedule.FinalSchedule == nil || len(schedule.FinalSchedule.RenderedScheduleEntries) == 0 {
		fmt.Println()
		fmt.Println(dimStyle.Render("  No on-call entries found in this date range."))
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println(headerStyle.Render("  On-call schedule for " + schedule.Name + ":"))
	printAllEntries(schedule.FinalSchedule.RenderedScheduleEntries)

	return nil
}
