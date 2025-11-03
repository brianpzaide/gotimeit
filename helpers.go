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
	return "", ErrEndSession
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

func transformActiveSessionsToWeekActivitiesSlice([]ActivitySession) []*WeekActivities {

	wa := make([]*WeekActivities, 0)

	return wa

}
