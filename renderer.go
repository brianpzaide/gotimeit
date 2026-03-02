package main

import (
	"bytes"
	"fmt"
	"strings"
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
		"formatDate": func(t string) string {
			tp, _ := time.Parse("2006-01-02", t)
			return tp.Format("Jan 02, 2006")
		},
		"upper": strings.ToUpper,
		"rowStart": func(index, weekDay int) int {
			return (index+weekDay)%7 + 1
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
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;
        background: #f5f7fa;
        padding: 40px 20px;
      }

      .container {
        max-width: 1100px;
        margin: 0 auto;
        display: flex;
        flex-direction: column;
        gap: 30px;
        align-items: center;
      }

      .card {
        background: #ffffff;
        padding: 28px 32px;
        border-radius: 12px;
        box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
        max-width: 1100px;
        width: 100%;
        text-align: center;    
        display: flex;
        flex-direction: column;
        align-items: center;   
      }

      .card h2 {
        margin-top: 0;
        margin-bottom: 20px;
        font-size: 20px;
        font-weight: 600;
        color: #222;
      }

      .heatmap {
        display: flex;
        gap: 22px;
        justify-content: center; 
        overflow-x: auto;
        padding-bottom: 10px;
      }

      .month {
        display: flex;
        flex-direction: column;
        align-items: center;
      }

      .month-label {
        font-size: 12px;
        font-weight: 600;
        margin-bottom: 6px;
        color: #444;
      }

      .month-grid {
        display: grid;
        grid-template-rows: repeat(7, 9px);
        grid-auto-flow: column;
        grid-auto-columns: 9px;
        gap: 3px;
      }

      .day {
        width: 9px;
        height: 9px;
        border-radius: 2px;
        cursor: pointer;
      }

      .level-0 { background-color: #ebedf0; }
      .level-1 { background-color: #c6e48b; }
      .level-2 { background-color: #7bc96f; }
      .level-3 { background-color: #239a3b; }
      .level-4 { background-color: #196127; }

      .tooltip {
        position: fixed;
        pointer-events: none;
        background: #333;
        color: white;
        font-size: 11px;
        padding: 6px 8px;
        border-radius: 4px;
        white-space: nowrap;
        z-index: 9999;
        display: none;
        max-width: 220px;
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

      /* Session card */
      .small-card {
        max-width: 480px;
      }

      .instruction {
        font-size: 15px;
        color: #555;
        margin-bottom: 15px;
        line-height: 1.4;
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
        border-color: #007bff;
        outline: none;
      }

      .stop-btn {
        background-color: #dc3545;
        width: 100%;
        margin-top: 10px;
      }

      .stop-btn:hover {
        background-color: #b02a37;
      }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js"></script>
  </head>	

  <body>
    <div class="container">
      <form hx-get="/summary" hx-trigger="submit" hx-target="#activity-chart">
        <select name="year"> 
        	{{range .YearOptions}}
    	  	  <option value="{{.}}">{{.}}</option>
    	    {{end}}
        </select>
        <button type="submit">Submit</button> 
      </form>
      <div class="card" id="activity-chart">
        {{with .CurrentYearActivityChartData}}
          <h2>Activity Tracker for {{ .Year }}</h2>
          <div class="heatmap">
            {{ range $month, $monthData := .MonthDailyActivities }}
                <div class="month">
                  <div class="month-label">{{ $month }}</div>
                  <div class="month-grid">
                    {{ range $index, $dayActivities := $monthData.DA }}
                      <div class="day level-{{ $dayActivities.Level }}"
                           data-tooltip='
                             <strong>{{ formatDate $dayActivities.Date }}</strong><br>
                             {{range $activity, $hours := $dayActivities.Activities}}
                               {{$activity}}: {{$hours}} hrs<br>
                             {{end}}
                           ' style="grid-row-start: {{ rowStart $monthData.Offset  $index }};">
                      </div>
                    {{ end }}
                  </div>
                </div>
            {{ end }}
          </div>
        {{end}}
      </div>

      <div class="card small-card">
        <div id="session-action">
          {{if .ActiveSession}} 
            <div class="instruction">Session for the activity <strong>{{.ActiveSession | upper}}</strong> is currently active. To start a new session click Stop first to end the current session</div>
            <form hx-get="/sessions/end" hx-trigger="submit" hx-target="#session-action">
              <button type="submit" style="background-color: red; width: 100%; margin-top: 9px;">
                Stop
              </button>
            </form>
          {{else}} 
            <div class="instruction">Enter your activity name below and click Start to begin a new session.</div>
            <form hx-get="/sessions/start" hx-trigger="submit" hx-target="#session-action">
              <input type="text" name="activity" id="activity" placeholder="Enter activity name" required>
              <button type="submit">Start Session</button>
            </form>
          {{end}}
        </div>
      </div>
    </div>

    <div id="tooltip" class="tooltip"></div>

    <script>
      const tooltip = document.getElementById("tooltip");
      
      document.addEventListener("mouseover", function (e) {
        const day = e.target.closest(".day");
        if (!day) return;
      
        tooltip.innerHTML = day.dataset.tooltip;
        tooltip.style.display = "block";
      });

      document.addEventListener("mousemove", function (e) {
        if (tooltip.style.display !== "block") return;

        const padding = 12;
        let x = e.clientX + padding;
        let y = e.clientY + padding;

        if (x + tooltip.offsetWidth > window.innerWidth) {
          x = e.clientX - tooltip.offsetWidth - padding;
        }

        if (y + tooltip.offsetHeight > window.innerHeight) {
          y = e.clientY - tooltip.offsetHeight - padding;
        }

        tooltip.style.left = x + "px";
        tooltip.style.top = y + "px";
      });

      document.addEventListener("mouseout", function (e) {
        if (e.target.closest(".day")) {
          tooltip.style.display = "none";
        }
      });
    </script>


  </body>
</html>
`

const ACTIVITY_CHART_HTML = `
<h2>Activity Tracker for {{ .Year }}</h2>
<div class="heatmap">
  {{ range $month, $monthData := .MonthDailyActivities }}
      <div class="month">
        <div class="month-label">{{ $month }}</div>
        <div class="month-grid">
          {{ range $index, $dayActivities := $monthData.DA }}
            <div class="day level-{{ $dayActivities.Level }}"
                 data-tooltip='
                   <strong>{{ formatDate $dayActivities.Date }}</strong><br>
                   {{range $activity, $hours := $dayActivities.Activities}}
                     {{$activity}}: {{$hours}} hrs<br>
                   {{end}}
                 ' style="grid-row-start: {{ rowStart $monthData.Offset  $index }};">
            </div>
          {{ end }}
        </div>
      </div>
  {{ end }}
</div>
`

const END_ACTIVITY_HTML = `
<div class="instruction">Session for the activity <strong>{{.ActiveSession | upper}}</strong> is currently active. To start a new session click Stop first to end the current session</div>
<form hx-get="/sessions/end" hx-trigger="submit" hx-target="#session-action">
  <button type="submit" style="background-color: red; width: 100%; margin-top: 9px;">
    Stop
  </button>
</form>
`
const START_ACTIVITY_HTML = `
<div class="instruction">Enter your activity name below and click Start to begin a new session.</div>
<form hx-get="/sessions/start" hx-trigger="submit" hx-target="#session-action">
  <input type="text" name="activity" id="activity" placeholder="Enter activity name" required>
  <button type="submit">Start Session</button>
</form>
`
const NO_ACTIVITY_DATA_FOUND_HTML = `
<div class="instruction">No activity records found for the year {{.Year}}.</div>
`
