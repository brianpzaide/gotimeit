import sqlite3
import random

DEFAULT_DB_FILE_PATH = "./activitysessions.db"
DEFAULT_SCHEMA_FILE_PATH = "./schema.sql"

YEARS = [2020,2021,2022, 2023, 2024, 2025]
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
            for c in ['programming', 'writing', 'reading', 'volunteering']:
                for _ in range(random.randint(3, 10)):
                    buy = random.randint(0,100)
                    sell = random.randint(buy, buy+25)
                    cursor.execute(insert_item, (f"{y}-{m:02}-{d:02}", c, buy, sell))
    
    connection.commit()
    connection.close()


# test out the following queries against the database to ensure their correctness
"""
SELECT strftime('%m', date) AS month, activity, SUM(stop_time-start_time) as hours FROM activitysessions WHERE stop_time is NOT NULL AND strftime('%Y', date) = ? GROUP BY month, activity;

SELECT strftime('%Y', date) AS year, activity, SUM(stop_time-start_time) as hours FROM activitysessions WHERE stop_time is NOT NULL GROUP BY year, activity;
"""

if __name__ == '__main__':
    init_database()
    main()