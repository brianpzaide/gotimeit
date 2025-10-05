package main

import "errors"

const ErrStartSession = "A session for %s is still active. Please end it before starting a new one"

var ErrEndSession = errors.New("no current session in progress")

type ActivitySessionInfo struct {
	Id        int
	Activity  string
	StartTime int64
}

type ActivitiesSessionsToday struct {
	Durations  []float32 `json:"series"`
	Activities []string  `json:"labels"`
}

type ActivitySession struct {
	Activity string
	Duration float32
}

type MonthActivitySession struct {
	ActivitySession
	Month int
}

type YearActivitySession struct {
	ActivitySession
	Year int
}

type ActivitySessions struct {
	// "Activity" represents name of the activity such as "programming", "writing" etc
	Activity string `json:"name"`
	// represents the amount of time spent each month
	Sessions []float32 `json:"data"`
}

type CurrentYearActivitySessions struct {
	// represents the "title" of the stacked bar chart
	Title string `json:"title"`
	// represents name of the activity and the amount of time spent each month, for example. [
	// {Activity:"programming", Sessions: [1,2,3,4,5,6,7,8,9,10,11,12]},
	// {Activity:"writing", Sessions: [1,2,3,4,5,6,7,8,9,10,11,12]},
	// ]
	ActivitySessions []ActivitySessions `json:"series"`
}

type Stroke struct {
	Width []int    `json:"width"`
	Curve []string `json:"curve"`
}

type OverAllYearsActivitySessions struct {
	// represents years for example ["2020", "2021", "2022"], this name is chosen because the apex charts expects this name
	Catagories []int `json:"catagories"`
	// represents name of the activity and the amount of time spent each month, for example. [
	// {Activity:"programming", Sessions: [1,2,3,4,5,6,7,8,9,10,11,12]},
	// {Activity:"writing", Sessions: [1,2,3,4,5,6,7,8,9,10,11,12]},
	// ]
	ActivitySessions []ActivitySessions `json:"series"`
	Stroke           Stroke             `json:"stroke"`
}

type templateData struct {
	TodaysData                   ActivitiesSessionsToday      `json:"todays_data"`
	CurrentYearMonthlyData       CurrentYearActivitySessions  `json:"monthly_data"`
	OverTheYearsActivitySessions OverAllYearsActivitySessions `json:"overall_data"`
}

type templateDataEnvelope struct {
	ActiveSession     string
	FlashErrorMessage string
	TmplDataJSON      []byte
}
