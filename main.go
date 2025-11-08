package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "Time Tracking CLI",
		Usage: "A simple CLI to measure time spent on hobbies",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Starts a new work session",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "activity",
						Usage:    "Name of the activity being tracked",
						Required: true,
					},
				},
				Action: handleStartSession,
			},

			{
				Name:   "end",
				Usage:  "Ends the current work session",
				Action: handleEndSession,
			},

			{
				Name:   "today",
				Usage:  "Displays the total hours spent on each activity for the current day in a tabular format",
				Action: handleTodaysSummary,
			},

			{
				Name:   "summary",
				Usage:  "Generates an interactive HTML summary with graphs. Starts a web server on port 4000 to view and manage sessions",
				Action: handleSummary,
			},
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
