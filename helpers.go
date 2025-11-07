package main

import (
	"errors"
	"fmt"
	"text/template"
	"time"
)

func startSession(activityName string) error {
	currSessionInfo, err := getCurrentSession()
	if err != nil {
		return fmt.Errorf("error fetching currently active session info: %v", err)
	}

	if currSessionInfo.Id == 0 {
		err = createActivitySession(activityName)
		if err != nil {
			return fmt.Errorf("error creating new session :%v", err)
		}
		return nil
	}

	return errors.New(ErrStartSession)
}

func endCurrentActiveSession() (string, error) {
	currSessionInfo, err := getCurrentSession()
	if err != nil {
		return "", fmt.Errorf("error fetching currently active session info: %v", err)
	}
	if currSessionInfo.Id != 0 {
		err = endActivitySession(currSessionInfo.Id)
		if err != nil {
			return "", fmt.Errorf("error ending an active session :%v", err)
		}
		return currSessionInfo.Activity, nil
	}
	return "", errors.New(ErrEndSession)
}

func todaysSummary() ([]ActivitySession, error) {
	todaysSessions, err := getTimeSpentOnEachActivityForToday()
	if err != nil {
		return nil, fmt.Errorf("error fetching activity sessions for today: %v", err)
	}

	return todaysSessions, nil
}

func computeTemplateData() error {
	if tHomepage == nil {
		tpl := template.Must(template.New("homepage").Funcs(funcMap).Parse(HOME_PAGE_HTML))
		tHomepage = tpl
	}

	yearOptions, err := getYearsOptions()
	if err != nil {
		return err
	}

	tmplData = &TemplateData{
		YearOptions: yearOptions,
	}

	// compute current years chart data
	currentYear := fmt.Sprintf("%d", time.Now().Year())
	chartData, err := computeChartDataForYear(currentYear)
	if err != nil {
		return err
	}
	tmplData.CurrentYearActivityChartData = chartData
	// err = updateTemplateData(false)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func computeChartDataForYear(year string) (*ActivityChartData, error) {

	// given year compute the chart data for that year and return it

	return nil, nil
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
				Date: "",
			}
		}
		days = append(days, da)
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
					Name:        monthName,
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
