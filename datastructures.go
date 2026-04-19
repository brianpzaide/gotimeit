package main

import "time"

const ErrStartSession = "a session is already in progress. Please end the current session before starting a new one"

var ErrEndSession = "no current session in progress"

type ActivitySession struct {
	Date     string
	Activity string
	Duration float32
}

type DayActivities struct {
	Date       string
	Activities map[string]float32
	TotalHours float32
	Level      int
}

type ActivityChartData struct {
	// stores info for the entire year's (52/53 week) activities grouped by months
	MonthDailyActivities map[time.Month]struct {
		Offset int
		DA     []*DayActivities
	}
	// for rendering the heading for the chart
	Year        string
	YearOptions []string
}

type TemplateData struct {
	// activity in current active session
	ActiveSession                string
	CurrentYearActivityChartData *ActivityChartData
}

type Segment struct {
	Activity string `json:"activity"`
	Start    int64  `json:"start"`
	End      int64  `json:"end"`
}
