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
	err := startSession(activityName)
	if err != nil {
		return err
	}
	fmt.Printf("New session for the activity %s has now started", activityName)
	return nil
}

func handleEndSession(ctx context.Context, c *cli.Command) error {
	activityName, err := endCurrentActiveSession()
	if err != nil {
		return err
	}
	fmt.Printf("Session for the activity %s has now ended", activityName)
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

	unknown := float32(24)
	for i, session := range todaysSessions {
		r := []*simpletable.Cell{
			{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", i+1)},
			{Text: session.Activity},
			{Align: simpletable.AlignRight, Text: fmt.Sprintf("$ %.2f", session.Duration)},
		}
		table.Body.Cells = append(table.Body.Cells, r)
		unknown -= session.Duration
	}

	// this is the time spent on untracked activities
	r := []*simpletable.Cell{
		{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", len(todaysSessions)+1)},
		{Text: "unknown"},
		{Align: simpletable.AlignRight, Text: fmt.Sprintf("$ %.2f", unknown)},
	}
	table.Body.Cells = append(table.Body.Cells, r)

	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
	return nil
}

func handleSummary(ctx context.Context, c *cli.Command) error {
	err := computeTemplateData()
	if err != nil {
		return fmt.Errorf("error computing the template data used for server side rendering: %v", err)
	}
	err = serve()
	if err != nil {
		return fmt.Errorf("error running the server: %v", err)
	}
	return nil
}
