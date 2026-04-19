package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

var months = map[time.Month]int{
	time.January:   31,
	time.February:  28,
	time.March:     31,
	time.April:     30,
	time.May:       31,
	time.June:      30,
	time.July:      31,
	time.August:    31,
	time.September: 30,
	time.October:   31,
	time.November:  30,
	time.December:  31,
}

func startSession(activityName string) (string, error) {
	activeSessionActivity, err := createActivitySession(activityName)
	if err != nil {
		if err.Error() == ErrStartSession {
			return activeSessionActivity, err
		}
		return "", err
	}

	return "", nil
}

func endCurrentActiveSession() (string, string, error) {
	date, activity, err := endActivitySession()
	if err != nil {
		return "", "", err
	}

	return date, activity, nil
}

func todaysSummary() ([]ActivitySession, error) {
	// sqlite understands ISO format yyyy-mm-dd
	today := time.Now().Format("2006-01-02")
	todaysSessions, err := getTimeSpentOnEachActivityFor(today)
	if err != nil {
		return nil, fmt.Errorf("error fetching activity sessions for today: %v", err)
	}

	return todaysSessions, nil
}

func computeTemplateData() (*TemplateData, error) {
	tmplData := &TemplateData{
		// YearOptions: yearOptions,
	}
	as, err := getCurrentActiveSession()
	if err != nil {
		return nil, err
	}
	tmplData.ActiveSession = as
	mu.Lock()
	defer mu.Unlock()
	chartData, OK := chartDataByYear[currentYear]
	if !OK {
		cd, err := computeChartDataForYear(currentYear)
		if err != nil {
			return nil, err
		}
		chartDataByYear[currentYear] = cd
		chartData = cd
	}
	tmplData.CurrentYearActivityChartData = chartData
	return tmplData, nil
}

func computeChartDataForYear(year string) (*ActivityChartData, error) {
	as, err := getTimeSpentOnEachActivityEverydayForYear(year)
	if err != nil {
		return nil, err
	}
	y, err := strconv.Atoi(year)
	if err != nil {
		return nil, err
	}
	//
	activityChartData := transformActiveSessionsToActivityChartData(y, as)
	activityChartData.YearOptions = yearOptions
	return activityChartData, nil
}

func getLevel(k float32) int {
	if k == float32(0) {
		return 0
	}
	if k < float32(2) {
		return 1
	}
	if k < float32(4) {
		return 2
	}
	if k < float32(6) {
		return 3
	}
	return 4
}

func isLeapYear(year int) bool {
	if year%4 == 0 {
		if year%100 == 0 {
			return year%400 == 0
		}
		return true
	}
	return false
}

func updateChartDataForCurrentYear(date string) error {
	mu.Lock()
	defer mu.Unlock()
	cy := fmt.Sprintf("%d", time.Now().Year())
	mwam, OK := chartDataByYear[cy]
	if !OK {
		mwam, err := computeChartDataForYear(cy)
		if err != nil {
			return err
		}
		chartDataByYear[cy] = mwam
		return nil
	}

	todaysSummary, err := getTimeSpentOnEachActivityFor(date)
	if err != nil {
		return err
	}
	activities := make(map[string]float32)
	totalHours := float32(0)
	for _, activitySession := range todaysSummary {
		activities[activitySession.Activity] = activitySession.Duration
		totalHours += activitySession.Duration
	}

	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Fatal(err)
	}
	month, dayNumber := t.Month(), t.Day()-1

	mwam.MonthDailyActivities[month].DA[dayNumber].Date = date
	mwam.MonthDailyActivities[month].DA[dayNumber].Activities = activities
	mwam.MonthDailyActivities[month].DA[dayNumber].TotalHours = totalHours
	mwam.MonthDailyActivities[month].DA[dayNumber].Level = getLevel(totalHours)

	chartDataByYear[cy] = mwam

	return nil
}

func transformActiveSessionsToActivityChartData(year int, activitySessions []ActivitySession) *ActivityChartData {
	daMap := make(map[string]*DayActivities)

	for _, as := range activitySessions {
		da, OK := daMap[as.Date]
		if !OK {
			da = &DayActivities{
				Date:       as.Date,
				Activities: make(map[string]float32),
				TotalHours: 0,
				Level:      0,
			}
			daMap[as.Date] = da
		}
		da.Activities[as.Activity] = as.Duration
		da.TotalHours += as.Duration
		da.Level = getLevel(da.TotalHours)
	}

	monthDailyActivitiesMap := make(map[time.Month]struct {
		Offset int
		DA     []*DayActivities
	})

	for month, lastDay := range months {
		days := make([]*DayActivities, 0)
		ld := lastDay
		if isLeapYear(year) {
			ld += 1
		}
		start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year, month, ld, 0, 0, 0, 0, time.UTC)
		offset := int(start.Weekday())
		current := start

		// appending actual days of the year
		for current.Before(end) || current.Equal(end) {
			dateStr := current.Format("2006-01-02")
			da, OK := daMap[dateStr]
			if !OK {
				da = &DayActivities{
					Date:       dateStr,
					TotalHours: 0,
					Level:      getLevel(0),
				}
			}
			days = append(days, da)
			// increment the current by one day
			current = current.Add(24 * time.Hour)
		}

		monthDailyActivitiesMap[month] = struct {
			Offset int
			DA     []*DayActivities
		}{
			Offset: offset,
			DA:     days,
		}
	}
	mwam := &ActivityChartData{
		Year:                 fmt.Sprintf("%d", year),
		MonthDailyActivities: monthDailyActivitiesMap,
	}
	return mwam
}

func initializeTemplates() {
	// initialize all the homepage template
	if tHomepage == nil {
		tpl := template.Must(template.New("homepage").Funcs(funcMap).Parse(HOME_PAGE_HTML))
		tHomepage = tpl
	}

	// initialize all the chart template
	if tChart == nil {
		tpl := template.Must(template.New("chart").Funcs(funcMap).Parse(ACTIVITY_CHART_HTML))
		tChart = tpl
	}

	// initialize all the chart404 template
	if tChart404 == nil {
		tpl := template.Must(template.New("chart404").Parse(NO_ACTIVITY_DATA_FOUND_HTML))
		tChart404 = tpl
	}

	// initialize all the end session template
	if tEndSessionAction == nil {
		tpl := template.Must(template.New("endSession").Funcs(funcMap).Parse(END_ACTIVITY_HTML))
		tEndSessionAction = tpl
	}

	// initialize all the start session template
	if tStartSessionAction == nil {
		tpl := template.Must(template.New("startSession").Parse(START_ACTIVITY_HTML))
		tStartSessionAction = tpl
	}
}

type envelope map[string]interface{}

func writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
