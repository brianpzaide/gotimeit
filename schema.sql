DROP TABLE IF EXISTS activitysessions;

CREATE TABLE IF NOT EXISTS activitysessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    activity TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
	stop_time TIMESTAMP
);