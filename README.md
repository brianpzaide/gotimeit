[WIP]

# Gotimeit

Gotimeit is a simple CLI tool to track how much time you spend on activities like programming, writing, or any other skill you are working to improve. It helps you quantify your actual time spent on that activity.

### Database Model:

Gotimeit uses a single table: activitysessions. Each session has:
* ```id```: Primary key
* ```date```: Date when the session started
* ```activity```: Name of the activity (e.g., programming, writing)
* ```start_time```: When the session started
* ```stop_time```: When the session ended

### Commands

* ```start```: Start tracking time for an activity. 
```bash
gotimeit start --activity writing
```

* ```end```: End the current session.
```bash
gotimeit end
```

* ```today```: See how much time you've spent today, per activity.
```bash
gotimeit today
```

* ```summary```:  Starts a local web server with charts and stats.
    * Opens a dashboard on localhost:4000
    * Includes:
        * Today's activity breakdown
        * Monthly and yearly stats
        * Line chart of activity over time
        * Ability to start/stop sessions from the UI
```bash
gotimeit summary
```

## Build and run the CLI

```bash
go build -o gotimeit .

./gotimeit summary
```

This starts the web dashboard on `http://localhost:4000`.

For testing or demo purposes, you can use `fakedatagen.py` to generate fake data. 


## Credits

gotimeit uses on the following libraries
* [urfave/cli](https://github.com/urfave/cli/v3) - CLI framework
* [chi](https://github.com/go-chi/chi/) - HTTP router
* [ApexCharts](https://github.com/apexcharts) - Frontend charts
* [SimpleTable](https://github.com/alexeyco/simpletable) - Terminal tables
