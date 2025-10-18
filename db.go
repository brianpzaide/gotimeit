package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const DSN = "./activitysessions.db"

const create_activitysessions_table = `CREATE TABLE IF NOT EXISTS activitysessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    activity TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
	stop_time TIMESTAMP
);`

const create_activity_session = `INSERT INTO activitysessions(date, activity, start_time) VALUES(?, ?, ?);`

const end_current_activity = `UPDATE activitysessions SET stop_time = ? where id = ?;`

const get_current_activity_session = `SELECT id, activity 
	FROM activitysessions 
	where stop_time is NULL 
	ORDER BY start_time DESC 
	LIMIT 1;`

const get_activity_sessions_for_today = `
	SELECT activity, ROUND(SUM(stop_time-start_time)*1.0/3600, 2) as hours 
	FROM activitysessions 
	WHERE date = ? AND stop_time is NOT NULL 
	GROUP BY activity;`

const get_activity_sessions_for_current_year = `
	SELECT strftime('%m', date) AS month, activity, ROUND(SUM(stop_time-start_time)*1.0/3600, 2) as hours 
	FROM activitysessions 
	WHERE stop_time is NOT NULL AND strftime('%Y', date) = ? 
	GROUP BY month, activity;`

const get_activity_sessions_over_the_years = `
	SELECT strftime('%Y', date) AS year, activity, ROUND(SUM(stop_time-start_time)*1.0/3600, 2) as hours 
	FROM activitysessions 
	WHERE stop_time is NOT NULL 
	GROUP BY year, activity
	ORDER BY year;`

func getDBConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", DSN)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func New() error {
	db, err := getDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(create_activitysessions_table)
	if err != nil {
		return err
	}
	return nil
}

func createActivitySession(activity string) error {
	db, err := getDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()
	now := time.Now()
	_, err = db.Exec(create_activity_session, now.Format("2006-01-02"), activity, now.Unix())
	return err
}

func getCurrentSession() (ActivitySessionInfo, error) {
	db, err := getDBConnection()
	if err != nil {
		return ActivitySessionInfo{}, err
	}
	defer db.Close()
	var currentSessionInfo ActivitySessionInfo
	err = db.QueryRow(get_current_activity_session).Scan(
		&currentSessionInfo.Id,
		&currentSessionInfo.Activity,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return ActivitySessionInfo{}, nil
		}
		return ActivitySessionInfo{}, err
	}

	return currentSessionInfo, nil
}

func endActivitySession(id int) error {
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

func getTimeSpentOnEachActivityForToday() ([]ActivitySession, error) {
	db, err := getDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// sqlite understands ISO format yyyy-mm-dd
	today := time.Now().Format("2006-01-02")

	rows, err := db.Query(get_activity_sessions_for_today, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := make([]ActivitySession, 0)

	i := 0
	for rows.Next() {
		var activitySession ActivitySession
		err = rows.Scan(
			&activitySession.Activity,
			&activitySession.Duration,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, activitySession)
		i += 1
	}
	return sessions, nil
}

func getTimeSpentOnEachActivityMonthly() ([]MonthActivitySession, error) {
	db, err := getDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	year := fmt.Sprintf("%d", time.Now().Year())
	rows, err := db.Query(get_activity_sessions_for_current_year, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	monthsActivitiesSessions := make([]MonthActivitySession, 0)
	for rows.Next() {
		var monthActivitySession MonthActivitySession
		err = rows.Scan(
			&monthActivitySession.Month,
			&monthActivitySession.Activity,
			&monthActivitySession.Duration,
		)
		if err != nil {
			return nil, err
		}
		monthsActivitiesSessions = append(monthsActivitiesSessions, monthActivitySession)
	}

	return monthsActivitiesSessions, nil
}

func getTimeSpentOnEachActivityOverTheYears() ([]YearActivitySession, error) {
	db, err := getDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(get_activity_sessions_over_the_years)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	yearsActivitiesSessions := make([]YearActivitySession, 0)
	for rows.Next() {
		var yearActivitySession YearActivitySession
		err = rows.Scan(
			&yearActivitySession.Year,
			&yearActivitySession.Activity,
			&yearActivitySession.Duration,
		)
		if err != nil {
			return nil, err
		}
		yearsActivitiesSessions = append(yearsActivitiesSessions, yearActivitySession)
	}

	return yearsActivitiesSessions, nil
}
