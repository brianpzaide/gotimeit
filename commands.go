package main

import (
	"context"
	"fmt"

	"github.com/alexeyco/simpletable"
	"github.com/urfave/cli/v3"
)

const DEFAULT_ACTIVITY = "programming"

func handleStartSession(ctx context.Context, c *cli.Command) error {
	activityName := c.String("activity")
	if activityName == "" {
		activityName = DEFAULT_ACTIVITY
	}
	activeSessionActivity, err := startSession(activityName)
	if err != nil {
		if err.Error() == ErrStartSession {
			fmt.Printf("Session for the activity %s is currently active. To start a new session end the current session first\n", activeSessionActivity)
		}
		return err
	}
	fmt.Printf("New session for the activity %s has now started\n", activityName)
	return nil
}

func handleEndSession(ctx context.Context, c *cli.Command) error {
	_, activity, err := endCurrentActiveSession()
	if err != nil {
		if err.Error() == ErrEndSession {
			fmt.Println(ErrEndSession)
		}
		return err
	}
	fmt.Printf("Session for the activity %s has now ended\n", activity)
	return nil
}

func handleTodaysSummary(ctx context.Context, c *cli.Command) error {
	todaysSessions, err := todaysSummary()
	if err != nil {
		return err
	}
	// outputing the results using alexeyco/simpletable
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "#"},
			{Align: simpletable.AlignCenter, Text: "NAME"},
			{Align: simpletable.AlignCenter, Text: "HOURS"},
		},
	}

	unTracked := float32(24)
	for i, session := range todaysSessions {
		r := []*simpletable.Cell{
			{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", i+1)},
			{Text: session.Activity},
			{Align: simpletable.AlignRight, Text: fmt.Sprintf("%.2f", session.Duration)},
		}
		table.Body.Cells = append(table.Body.Cells, r)
		unTracked -= session.Duration
	}

	// this is the time spent on untracked activities
	r := []*simpletable.Cell{
		{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", len(todaysSessions)+1)},
		{Text: "unTracked"},
		{Align: simpletable.AlignRight, Text: fmt.Sprintf("%.2f", unTracked)},
	}
	table.Body.Cells = append(table.Body.Cells, r)

	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
	return nil
}

func handleSummary(ctx context.Context, c *cli.Command) error {
	initializeTemplates()

	err := setYearsOptions()
	if err != nil {
		return err
	}

	// run the server
	err = serve()
	if err != nil {
		return fmt.Errorf("error running the server: %v", err)
	}
	return nil
}
