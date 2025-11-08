package main

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"
	"time"
)

const weekWidthPixel = 15

var (
	tHomepage           *template.Template
	tChart              *template.Template
	tChart404           *template.Template
	tStartSessionAction *template.Template
	tEndSessionAction   *template.Template
	mu                  *sync.Mutex = &sync.Mutex{}
	currentYear         string      = fmt.Sprintf("%d", time.Now().Year())
	yearOptions         []string
	chartDataByYear     map[string]*ActivityChartData = make(map[string]*ActivityChartData)
	funcMap             map[string]any                = template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 02, 2006")
		},
	}
)

func renderHomepage(tmplData *TemplateData) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := tHomepage.Execute(buf, tmplData)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderChart(acd *ActivityChartData) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := tChart.Execute(buf, acd)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderStartSessionAction() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := tStartSessionAction.Execute(buf, nil)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderEndSessionAction(activity string) ([]byte, error) {
	buf := new(bytes.Buffer)

	td := struct {
		ActiveSession string
	}{ActiveSession: activity}

	err := tEndSessionAction.Execute(buf, td)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderNoDataAvailable(year string) ([]byte, error) {
	buf := new(bytes.Buffer)

	td := struct {
		Year string
	}{Year: year}

	err := tStartSessionAction.Execute(buf, td)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const HOME_PAGE_HTML = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>GoTimeit</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      display: flex;
      flex-direction: column;
      align-items: center;
      background: #f8f9fa;
      margin-top: 20px;
    }

    h2 { margin-bottom: 10px; }

    .card {
      background: #fff;
      padding: 30px 25px;
      border-radius: 12px;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
      width: 100%;
      max-width: 500px;
      display: flex;
      flex-direction: column;
      margin: auto;
      align-items: center;
      justify-content: center;
      margin-bottom: 50px;
      gap: 20px;
    }

    .input-row {
      display: flex;
      gap: 10px;
    }

    input[type="text"] {
      flex: 1;
      padding: 10px 12px;
      border: 1px solid #ccc;
      border-radius: 8px;
      font-size: 16px;
      transition: border-color 0.3s;
    }

    input[type="text"]:focus {
      border-color: #a7a7a7;
      outline: none;
    }

    button {
      background-color: #007bff;
      color: white;
      border: none;
      padding: 10px 16px;
      border-radius: 8px;
      font-size: 16px;
      cursor: pointer;
      transition: background-color 0.3s;
    }
    
    button:hover {
      background-color: #0056b3;
    }
    
    .instruction {
      font-size: 15px;
      color: #444;
      line-height: 1.4;
    }

    .chart-container {
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      background: #fff;
      padding: 10px;
      border-radius: 8px;
      border: 1px solid #ddd;
      position: relative;
    }

    .months {
      display: flex;
      position: relative;
      margin-left: 20px;
      margin-bottom: 5px;
      height: 15px;
      font-size: 12px;
      color: #666;
    }

    .month-label {
      position: absolute;
      top: 0;
    }

    .contribution-chart {
      display: flex;
      flex-direction: row;
      gap: 3px;
    }

    .week {
      display: flex;
      flex-direction: column;
      gap: 3px;
    }

    .day {
      width: 12px;
      height: 12px;
      border-radius: 2px;
      background-color: #ebedf0;
      position: relative;
    }

    .level-0 { background-color: #d8ceceff; }
    .level-1 { background-color: #a6d4b6ff; }
    .level-2 { background-color: #71c792ff; }
    .level-3 { background-color: #3c995bff; }
    .level-4 { background-color: #137033ff; }

    .day:hover {
      outline: 1px solid #555;
      cursor: pointer;
      z-index: 2;
    }

    .tooltip {
      display: none;
      position: absolute;
      top: -5px;
      left: 20px;
      background: #333;
      color: white;
      font-size: 11px;
      padding: 6px;
      border-radius: 4px;
      white-space: nowrap;
      z-index: 10;
      transform: translateY(-50%);
    }

    .tooltip table {
      border-collapse: collapse;
    }

    .tooltip td {
      padding: 2px 6px;
    }

    .day:hover .tooltip {
      display: block;
    }
  </style>
</head>

<body>
  <form hx-get="/summary" hx-trigger="submit"> 
    <select name="year" hx-target='#activity-chart' hx-indicator=".htmx-indicator"> 
      {{range .YearOptions}}
        <option value="{{.}}">{{.}}</option>
      {{end}}
    </select>
    <button type="submit">Submit</button> 
  </form>
  
  <div style="display: flex; margin: auto; align-items: center; justify-content: center;">
    <div id="activity-chart">
      {{with .CurrentYearActivityChartData}}
        <h2>Activity Tracker for {{ .Year }}</h2>
        <div class="chart-container">
          <div class="months">
            {{range .MonthLabels}}
              <span class="month-label" style="left: {{ .PixelOffset }}px;">{{ .Name }}</span>
            {{end}}
          </div>

          <div class="contribution-chart">
            {{range .WeeklyActivities}}
              <div class="week">
                {{range .DayActivities}}
                  {{if .Date}}
                    <div class="day level-{{ .Level }}">
                      <div class="tooltip">
                        <strong>{{ formatDate .Date }}</strong>
                        <table>
                          {{range $activity, $hours := .Activities}}
                            <tr><td>{{ $activity }}</td><td>{{ $hours }} hrs</td></tr>
                          {{end}}
                        </table>
                      </div>
                    </div>
                  {{ else }}
                    <div class="day level-0"></div>
                  {{ end }}
                {{ end }}
              </div>
            {{ end }}
          </div>
        </div>
      {{end}}
    </div>
  </div>

  <div class="card">
    <div id="session-action">
      {{if .ActiveSession}} 
        
      {{else}} 
        
      {{end}}
    </div>
  </div>

</body>
`

const ACTIVITY_CHART_HTML = `
<h2>Activity Tracker for {{ .Year }}</h2>
<div class="chart-container">
  <div class="months">
    {{range .MonthLabels}}
      <span class="month-label" style="left: {{ .x }}px;">{{ .name }}</span>
    {{end}}
  </div>

  <div class="contribution-chart">
    {{range .WeeklyActivities}}
      <div class="week">
        {{range .DayActivities}}
          {{if .Date}}
            <div class="day level-{{ .Level }}">
              <div class="tooltip">
                <strong>{{ formatDate .Date }}</strong>
                <table>
                  {{range $activity, $hours := .Activities}}
                    <tr><td>{{ $activity }}</td><td>{{ $hours }} hrs</td></tr>
                  {{end}}
                </table>
              </div>
            </div>
          {{ else }}
            <div class="day level-0"></div>
          {{ end }}
        {{ end }}
      </div>
    {{ end }}
  </div>
</div>
`

const END_ACTIVITY_HTML = `
<div class="instruction">Session for the activity {{.ActiveSession}} is currently active. To start a new session click Stop first to end the current session</div>
<form hx-get="/sessions/end" hx-trigger="submit" hx-target="#session-action">
  <button type="submit" style="background-color: red; width: 100%; margin-top: 9px;">
    Stop
  </button>
</form>
`
const START_ACTIVITY_HTML = `
<div class="instruction">Enter your activity name below and click Start to begin a new session.</div>
<form hx-get="/sessions/start/{activity}" hx-trigger="submit" hx-target="#session-action">
  <input type="text" name="activity" id="activity" placeholder="Enter activity name" required>
  <button type="submit">Start Session</button>
</form>
`
const NO_ACTIVITY_DATA_FOUND_HTML = `
<div class="instruction">No activity records found for the year {{.Year}}.</div>
`
