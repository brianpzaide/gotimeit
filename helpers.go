package main

import (
	"errors"
	"fmt"
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

func transformActiveSessionsToWeekActivitiesSlice([]ActivitySession) ([]*WeekActivities, error) {

	return nil, nil

}
