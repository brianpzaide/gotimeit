package main

const ErrStartSession = "a session is already in progress. Please end the current session before starting a new one"

var ErrEndSession = "no current session in progress"

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
	Date     string
	Activity string
	Duration float32
}

type ActivitySessions struct {
	// "Activity" represents name of the activity such as "programming", "writing" etc
	Activity string `json:"name"`
	// represents the amount of time spent each month
	Sessions []float32 `json:"data"`
}

type MonthLabel struct {
	Name        string
	PixelOffset int
}

type DayActivities struct {
	Date       string
	Activities map[string]float32
	TotalHours float32
	Level      int
}

type WeekActivities struct {
	DailyActivities []*DayActivities
}

type ActivityChartData struct {
	// stores info for the entire year's (52/53 week) activities
	WeeklyActivities []*WeekActivities
	// for rendering the heading for the chart
	Year string
	// for displaying the months header in the chart
	MonthLabels []MonthLabel
}

type TemplateData struct {
	// activity in current active session
	ActiveSession                string
	YearOptions                  []string
	CurrentYearActivityChartData *ActivityChartData
}
