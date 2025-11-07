package main

import "time"

const test_start_session_query = `
WITH active_session AS (
    SELECT id, activity
    FROM activitysessions
    WHERE stop_time IS NULL
),
inserted AS (
    INSERT INTO activitysessions(date, activity, start_time)
    SELECT ?, ?, ?
    WHERE NOT EXISTS (SELECT 1 FROM active_session)
    RETURNING 1 AS result, '' AS activity
)
SELECT result, activity
FROM inserted
UNION ALL
SELECT -1 AS result, activity
FROM active_session
WHERE NOT EXISTS (SELECT 1 FROM inserted);`

const test_end_session_query = `
WITH active_session AS (
    SELECT date
    FROM activitysessions
    WHERE stop_time IS NULL
),
updated AS (
    UPDATE activitysessions
	SET stop_time = ?
    WHERE EXISTS (SELECT 1 FROM active_session)
    RETURNING 1 AS result, date
)
SELECT result, date
FROM updated
UNION ALL
SELECT -1 AS result, '' as date
FROM active_session
WHERE NOT EXISTS (SELECT 1 FROM updated);
`

func main() {
	// app := &cli.Command{
	// 	Name:  "Time Tracking CLI",
	// 	Usage: "A simple CLI to measure time spent on hobbies",
	// 	Commands: []*cli.Command{
	// 		{
	// 			Name:  "start",
	// 			Usage: "Starts a new work session",
	// 			Flags: []cli.Flag{
	// 				&cli.StringFlag{
	// 					Name:     "activity",
	// 					Usage:    "Name of the activity being tracked",
	// 					Required: true,
	// 				},
	// 			},
	// 			Action: handleStartSession,
	// 		},

	// 		{
	// 			Name:   "end",
	// 			Usage:  "Ends the current work session",
	// 			Action: handleEndSession,
	// 		},

	// 		{
	// 			Name:   "today",
	// 			Usage:  "Displays the total hours spent on each activity for the current day in a tabular format",
	// 			Action: handleTodaysSummary,
	// 		},

	// 		{
	// 			Name:   "summary",
	// 			Usage:  "Generates an interactive HTML summary with graphs. Starts a web server on port 4000 to view and manage sessions",
	// 			Action: handleSummary,
	// 		},
	// 	},
	// }

	// err := app.Run(context.Background(), os.Args)
	// if err != nil {
	// 	log.Fatal(err)
	// }

}

func testDBAction(id int) error {
	db, err := getDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(end_current_activity, time.Now().Unix(), id)
	if err != nil {
		return err
	}

	return nil
}
