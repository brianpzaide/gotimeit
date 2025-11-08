package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const DSN = "./activitysessions.db"

const start_session = `
	WITH active_session AS (
	    SELECT id, activity
	    FROM activitysessions
	    WHERE stop_time IS NULL
	),
	inserted AS (
	    INSERT INTO activitysessions(date, activity, start_time)
	    SELECT ?, ?, ?
	    WHERE NOT EXISTS (SELECT 1 FROM active_session)
	    RETURNING 1 AS result, '' AS activity
	)
	SELECT result, activity
	FROM inserted
	UNION ALL
	SELECT -1 AS result, activity
	FROM active_session
	WHERE NOT EXISTS (SELECT 1 FROM inserted);`

const end_session = `
	WITH active_session AS (
	    SELECT date, activity
	    FROM activitysessions
	    WHERE stop_time IS NULL
	),
	updated AS (
	    UPDATE activitysessions
		SET stop_time = ?
	    WHERE EXISTS (SELECT 1 FROM active_session)
	    RETURNING 1 AS result, date, activity
	)
	SELECT result, date, activity
	FROM updated
	UNION ALL
	SELECT -1 AS result, '' AS date, '' AS activity
	FROM active_session
	WHERE NOT EXISTS (SELECT 1 FROM updated);`

const create_activitysessions_table = `CREATE TABLE IF NOT EXISTS activitysessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    activity TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
	stop_time TIMESTAMP
);`

const get_activity_sessions_for_today = `
	SELECT activity, ROUND(SUM(stop_time-start_time)*1.0/3600, 2) as hours 
	FROM activitysessions 
	WHERE date = ? AND stop_time is NOT NULL 
	GROUP BY activity;`

const get_activity_sessions_everyday_for_year = `
	SELECT date, activity, ROUND(SUM(stop_time-start_time)*1.0/3600, 2) as hours 
	FROM activitysessions 
	WHERE strftime('%Y', date) = ? AND stop_time is NOT NULL
	GROUP BY date, activity
	ORDER BY date;`

const get_oldest_and_latest_years = `
	SELECT 
    (SELECT strftime('%Y', date) FROM activitysessions ORDER BY id ASC  LIMIT 1) AS oldest_year,
    (SELECT strftime('%Y', date) FROM activitysessions ORDER BY id DESC LIMIT 1) AS latest_year;`

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

func createActivitySession(activity string) (string, error) {
	db, err := getDBConnection()
	if err != nil {
		return "", err
	}
	defer db.Close()
	now := time.Now()
	row := db.QueryRow(start_session, now.Format("2006-01-02"), activity, now.Unix())
	isCreated, activeSessionActivity := 0, ""
	err = row.Scan(
		&isCreated,
		&activeSessionActivity,
	)
	if err != nil {
		return "", err
	}
	if isCreated == -1 {
		return activeSessionActivity, errors.New(ErrStartSession)
	}
	return "", nil

}

// func getCurrentSession() (ActivitySessionInfo, error) {
// 	db, err := getDBConnection()
// 	if err != nil {
// 		return ActivitySessionInfo{}, err
// 	}
// 	defer db.Close()
// 	var currentSessionInfo ActivitySessionInfo
// 	err = db.QueryRow(get_current_activity_session).Scan(
// 		&currentSessionInfo.Id,
// 		&currentSessionInfo.Activity,
// 		&currentSessionInfo.StartTime,
// 	)
// 	if err != nil {
// 		if err.Error() == "sql: no rows in result set" {
// 			return ActivitySessionInfo{}, nil
// 		}
// 		return ActivitySessionInfo{}, err
// 	}

// 	return currentSessionInfo, nil
// }

func endActivitySession() (string, string, error) {
	db, err := getDBConnection()
	if err != nil {
		return "", "", err
	}
	defer db.Close()
	now := time.Now()
	isStopped, date, activity := 0, "", ""
	row := db.QueryRow(end_session, now.Unix())

	err = row.Scan(
		&isStopped,
		&date,
		&activity,
	)
	if err != nil {
		return "", "", err
	}
	if isStopped == -1 {
		return "", "", errors.New(ErrEndSession)
	}
	return date, activity, nil
}

func getTimeSpentOnEachActivityFor(date string) ([]ActivitySession, error) {
	db, err := getDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(get_activity_sessions_for_today, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := make([]ActivitySession, 0)

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
	}
	return sessions, nil
}

func getTimeSpentOnEachActivityEverydayForYear(year string) ([]ActivitySession, error) {
	db, err := getDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(get_activity_sessions_everyday_for_year, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sessions := make([]ActivitySession, 0)

	for rows.Next() {
		var activitySession ActivitySession
		err = rows.Scan(
			&activitySession.Date,
			&activitySession.Activity,
			&activitySession.Duration,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, activitySession)
	}
	return sessions, nil
}

func setYearsOptions() error {
	db, err := getDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(get_oldest_and_latest_years)

	oldest, latest := 0, 0
	err = row.Scan(
		&oldest,
		&latest,
	)
	if err != nil {
		return err
	}
	yearOptions = make([]string, 0)
	if oldest == latest {
		yearOptions = append(yearOptions, fmt.Sprintf("%d", oldest))
		return nil
	}

	for i := oldest; i < latest; i++ {
		yearOptions = append(yearOptions, fmt.Sprintf("%d", i))
	}

	return nil
}
