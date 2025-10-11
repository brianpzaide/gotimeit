import sqlite3
import random

DEFAULT_DB_FILE_PATH = "./activitysessions.db"
DEFAULT_SCHEMA_FILE_PATH = "./schema.sql"

YEARS = [2020, 2021, 2022, 2023, 2024, 2025]
DAYS_IN_MONTHS = [31,28,31,30,31,30,31,31,30,31,30,31]

def init_database():
    connection = sqlite3.connect(DEFAULT_DB_FILE_PATH)
    with open(DEFAULT_SCHEMA_FILE_PATH, 'r') as f:
        connection.executescript(f.read())
    connection.commit()
    connection.close()


def main():
    insert_item = 'INSERT INTO activitysessions(date, activity, start_time, stop_time) Values(?,?,?,?);'
    connection = sqlite3.connect(DEFAULT_DB_FILE_PATH)
    cursor = connection.cursor()
    for y in YEARS:
        for (m, d) in enumerate(DAYS_IN_MONTHS, start=1):
            for i in range(1, d+1):
                beg = 0
                while beg < 24:
                    c  = random.choice(['programming', 'writing', 'reading', 'volunteering'])
                    dur = random.randint(0, 2)
                    if beg+dur < 24:
                        cursor.execute(insert_item, (f"{y}-{m:02}-{i:02}", c, beg*3600, (beg+dur)*3600))
                        beg += dur
                    else:
                        break
    
    connection.commit()
    connection.close()


# test out the following queries against the database to ensure their correctness
"""
SELECT strftime('%m', date) AS month, activity, SUM(stop_time-start_time) as hours FROM activitysessions WHERE stop_time is NOT NULL AND strftime('%Y', date) = ? GROUP BY month, activity;

SELECT strftime('%Y', date) AS year, activity, SUM(stop_time-start_time) as hours FROM activitysessions WHERE stop_time is NOT NULL GROUP BY year, activity;

SELECT activity, ROUND(SUM(stop_time-start_time)*1.0/3600, 2) as hours FROM activitysessions WHERE date = ? AND stop_time is NOT NULL GROUP BY activity;`

SELECT id, activity, start_time FROM activitysessions where stop_time is NULL ORDER BY start_time DESC LIMIT 1;

SELECT strftime('%m', date) AS month, activity, ROUND(SUM(stop_time-start_time)*1.0/3600, 2) as hours FROM activitysessions WHERE stop_time is NOT NULL AND strftime('%Y', date) = ? GROUP BY month, activity;

"""

if __name__ == '__main__':
    init_database()
    main()