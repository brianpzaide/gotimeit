package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const DSN = "./activitysessions.db"

const active_session = `SELECT activity FROM activitysessions WHERE stop_time IS NULL LIMIT 1`
const start_session = `INSERT INTO activitysessions(date, activity, start_time) VALUES (?, ?, ?)`

const end_session = `UPDATE activitysessions SET stop_time = ? WHERE stop_time IS NULL RETURNING date, activity`

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

func initializeDB() error {
	db, err := getDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(create_activitysessions_table)

	return err
}

func getCurrentActiveSession() (string, error) {
	db, err := getDBConnection()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var as sql.NullString
	err = db.QueryRow(active_session).Scan(&as)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return as.String, nil
}

func createActivitySession(activity string) (string, error) {
	inserted := true
	db, err := getDBConnection()
	if err != nil {
		return "", err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return "", err
	}

	var existingActivity sql.NullString
	err = tx.QueryRow(active_session).Scan(&existingActivity)

	switch {
	case err == sql.ErrNoRows:
		now := time.Now()
		_, err = tx.Exec(start_session, now.Format("2006-01-02"), activity, now.Unix())
		if err != nil {
			tx.Rollback()
			return "", err
		}

	default:
		inserted = false
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	if inserted {
		return activity, nil
	}

	return existingActivity.String, errors.New(ErrStartSession)
}

func endActivitySession() (string, string, error) {
	updated := true
	db, err := getDBConnection()
	if err != nil {
		return "", "", err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return "", "", err
	}

	var existingActivity, date, activity sql.NullString
	err = tx.QueryRow(active_session).Scan(&existingActivity)

	switch {
	case err == sql.ErrNoRows:
		updated = false

	default:
		now := time.Now()
		row := tx.QueryRow(end_session, now.Unix())
		row.Scan(&date, &activity)
		if err != nil {
			tx.Rollback()
			return "", "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}

	if updated {
		return date.String, activity.String, nil
	}
	return "", "", errors.New(ErrEndSession)
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
	fmt.Println("getTimeSpentOnEachActivityEverydayForYear: started")
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
	fmt.Printf("getTimeSpentOnEachActivityEverydayForYear: number of fetched rows %d\n", len(sessions))
	return sessions, nil
}

func setYearsOptions() error {
	db, err := getDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(get_oldest_and_latest_years)

	var oldestn, latestn sql.NullInt64
	var oldest, latest int
	err = row.Scan(
		&oldestn,
		&latestn,
	)
	if err != nil {
		return err
	}
	oldestn.Scan(oldest)
	latestn.Scan(latest)

	yearOptions = make([]string, 0)
	if oldest == 0 {
		yearOptions = append(yearOptions, fmt.Sprintf("%d", time.Now().Year()))
		return nil
	} else {
		if oldest == latest {
			yearOptions = append(yearOptions, fmt.Sprintf("%d", oldest))
			return nil
		}

		for i := oldest; i <= latest; i++ {
			yearOptions = append(yearOptions, fmt.Sprintf("%d", i))
		}

		return nil
	}

}
