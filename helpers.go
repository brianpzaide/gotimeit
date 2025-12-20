package main

import (
	"fmt"
	"log"
	"strconv"
	"text/template"
	"time"
)

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
		YearOptions: yearOptions,
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
	activityChartData := transformActiveSessionsToActivityChartData(y, as)

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

func updateChartDataForCurrentYear(date string) error {
	mu.Lock()
	defer mu.Unlock()
	cy := fmt.Sprintf("%d", time.Now().Year())
	acd, OK := chartDataByYear[cy]
	if !OK {
		acd, err := computeChartDataForYear(cy)
		if err != nil {
			return err
		}
		chartDataByYear[cy] = acd
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
	_, weekNumber := t.ISOWeek()
	dayNumber := int(t.Weekday())

	acd.WeeklyActivities[weekNumber].DailyActivities[dayNumber].Date = date
	acd.WeeklyActivities[weekNumber].DailyActivities[dayNumber].Activities = activities
	acd.WeeklyActivities[weekNumber].DailyActivities[dayNumber].TotalHours = totalHours
	acd.WeeklyActivities[weekNumber].DailyActivities[dayNumber].Level = getLevel(totalHours)

	chartDataByYear[cy] = acd

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

	days := make([]*DayActivities, 0)

	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	current := start

	// padding first week to align Jan 1 at the top-left
	for i := 0; i < int(start.Weekday()); i++ {
		da := &DayActivities{
			Date: "",
		}
		days = append(days, da)
	}

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

	// padding last week to align Dec 31 at the bottom-right
	for i := int(start.Weekday()); i < 7; i++ {
		da := &DayActivities{
			Date: "",
		}
		days = append(days, da)
	}

	// monthStarts and monthLabels are for displaying the month labels acurately as the chart header
	monthStarts := make(map[string]bool)
	monthLabels := make([]MonthLabel, 0)

	week, weeks := make([]*DayActivities, 0), make([]*WeekActivities, 0)

	for _, da := range days {
		if len(week) == 7 {
			wa := &WeekActivities{
				DailyActivities: week,
			}
			weeks = append(weeks, wa)
			week = make([]*DayActivities, 0)
		}
		week = append(week, da)
		if da.Date != "" {
			t, _ := time.Parse("2006-01-02", da.Date)
			monthName := t.Month().String()
			if _, OK := monthStarts[monthName]; !OK {
				ml := MonthLabel{
					Name:        monthName[:3],
					PixelOffset: weekWidthPixel * len(weeks),
				}
				monthLabels = append(monthLabels, ml)
				monthStarts[monthName] = true
			}
		}
	}

	acd := &ActivityChartData{
		Year:             fmt.Sprintf("%d", year),
		WeeklyActivities: weeks,
		MonthLabels:      monthLabels,
	}
	return acd
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
