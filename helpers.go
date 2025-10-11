package main

import (
	"fmt"
	"time"
)

// this function transforms the data coming from the database into something apex charts expects
// ([{Activity: "programming", "Duration"}, ...]) -> ActivitiesSessionsToday
func transformDataForTodaysSessions(todaysActivitiesSessions []ActivitySession) ActivitiesSessionsToday {
	activities, durations := make([]string, 0), make([]float32, 0)
	unTracked := float32(24)
	for _, tas := range todaysActivitiesSessions {
		activities = append(activities, tas.Activity)
		durations = append(durations, tas.Duration)
		unTracked -= tas.Duration
	}
	if unTracked > 0 {
		activities = append(activities, "UnTracked")
		durations = append(durations, unTracked)
	}
	return ActivitiesSessionsToday{
		Activities: activities,
		Durations:  durations,
	}
}

// this function transforms the data coming from the database into something apex charts expects
// ([{Activity: "programming", "Duration", Month: 1}, ...]) -> CurrentActivitySessions
func transformDataForCurrentYearSessions(monthsActivitiesSessions []MonthActivitySession) CurrentYearActivitySessions {
	fmt.Println("transformDataForCurrentYearSessions called")
	fmt.Println("transformDataForCurrentYearSessions data from db")
	for _, mas := range monthsActivitiesSessions {
		fmt.Printf("month: %d, activity: %s, duration: %.2f\n", mas.Month, mas.Activity, mas.Duration)
	}

	activitySessionsMap := make(map[string]*[12]float32)
	for _, mas := range monthsActivitiesSessions {
		if activitySessionsMap[mas.Activity] == nil {
			activitySessionsMap[mas.Activity] = &[12]float32{}
		}
		activitySessionsMap[mas.Activity][mas.Month-1] = mas.Duration
	}
	activitiesSessions := make([]ActivitySessions, 0)
	for activityName, sessions := range activitySessionsMap {
		activitiesSessions = append(activitiesSessions, ActivitySessions{
			Activity: activityName,
			Sessions: sessions[:],
		})
	}
	for _, as := range activitiesSessions {
		fmt.Println(as.Activity)
		for i, dur := range as.Sessions {
			fmt.Printf("month: %d, duration: %.2f\n", i, dur)
		}
	}
	return CurrentYearActivitySessions{
		Title:            fmt.Sprintf("Time spent on activities in %d", time.Now().Year()),
		ActivitySessions: activitiesSessions,
	}
}

// this function transforms the data coming from the database into something apex charts expects
// ([{Activity: "programming", "Duration", Year: 2020}, ...]) -> OverAllYearsActivitySessions
func transformDataForOverAllYearsSessions(yearsActivitiesSessions []YearActivitySession) OverAllYearsActivitySessions {
	activitySessionsMap := make(map[string][]float32)
	yearsRecorded := make(map[int]bool)
	for _, yas := range yearsActivitiesSessions {
		if activitySessionsMap[yas.Activity] == nil {
			activitySessionsMap[yas.Activity] = make([]float32, 0)
		}
		activitySessionsMap[yas.Activity] = append(activitySessionsMap[yas.Activity], yas.Duration)
		if _, ok := yearsRecorded[yas.Year]; !ok {
			yearsRecorded[yas.Year] = true
		}
	}
	activitiesSessions := make([]ActivitySessions, 0)
	for activityName, sessions := range activitySessionsMap {
		activitiesSessions = append(activitiesSessions, ActivitySessions{
			Activity: activityName,
			Sessions: sessions[:],
		})
	}

	catagories, width, curve := make([]int, 0), make([]int, 0), make([]string, 0)
	for k, _ := range yearsRecorded {
		catagories = append(catagories, k)
		width = append(width, 2)
		curve = append(curve, "straight")
	}
	stroke := Stroke{
		Width: width,
		Curve: curve,
	}
	return OverAllYearsActivitySessions{
		Catagories:       catagories,
		ActivitySessions: activitiesSessions,
		Stroke:           stroke,
	}
}

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

	return fmt.Errorf(ErrStartSession, currSessionInfo.Activity)
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

// func addErrMessageCookie(w http.ResponseWriter, msg string) {
// 	http.SetCookie(w, &http.Cookie{
// 		Name:  "flash",
// 		Value: url.QueryEscape(msg),
// 		Path:  "/",
// 	})
// }
