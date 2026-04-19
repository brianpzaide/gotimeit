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

// func renderNoDataAvailable(year string) ([]byte, error) {
// 	buf := new(bytes.Buffer)

// 	td := struct {
// 		Year string
// 	}{Year: year}

// 	err := tStartSessionAction.Execute(buf, td)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

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

      .segmentcard {
        max-width: 700px;
        margin: 40px auto;
        background: #fff;
        border-radius: 12px;
        padding: 20px 24px 28px;
        box-shadow: 0 6px 20px rgba(0,0,0,0.08);
      }

      .segmentcard-header {
        display: flex;
        justify-content: center;
        margin-bottom: 20px;
      }

      #datePicker {
        padding: 8px 12px;
        font-size: 14px;
        border-radius: 8px;
        border: 1px solid #ccc;
        background: #f9fafb;
        cursor: pointer;
      }

      .segmentcard-body {
        padding: 10px 0;
      }

      .bar {
        position: relative;
        width: 100%;
        height: 24px;
        background: #6a6a6a;
        border-radius: 12px;
        overflow: visible;
      }

      .segment {
        position: absolute;
        min-width: 2px;
        top: 0;
        height: 100%;
        background: #4caf50;
        z-index: 1;
      }

      .markers {
        position: absolute;
        width: 100%;
        height: 100%;
        pointer-events: none;
        z-index: 2;
      }

      .marker {
        position: absolute;
        top: 0;
        height: 100%;
        width: 1px;
        background: rgba(0,0,0,0.2);
      }

      .marker.major {
        background: rgba(0,0,0,0.4);
        width: 2px;
      }

      .segmenttooltip-global {
        position: fixed;
        top: 0;
        left: 0;
        background: #333;
        color: #fff;
        padding: 6px 10px;
        font-size: 12px;
        border-radius: 6px;
        white-space: nowrap;
        pointer-events: none;
        opacity: 0;
        transition: opacity 0.15s ease;
        z-index: 9999;
        box-shadow: 0 4px 12px rgba(0,0,0,0.2);
      }

      /* css for the heat map component*/
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

      /* css for the session card(start/stop) */
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
    <div id="segmenttooltip" class="segmenttooltip-global"></div>

    <div class="segmentcard">
      <div class="segmentcard-header">
        <input type="date" id="datePicker" />
      </div>
    
      <div class="segmanetcard-body">
        <div class="bar">
          <div class="markers" id="markers"></div>
        </div>
      </div>
    </div>

    <div class="container">
      <div class="card" id="activity-chart">
        <form hx-get="/summary" hx-trigger="submit" hx-target="#activity-chart">
          <select name="year"> 
          	{{range .CurrentYearActivityChartData.YearOptions}}
    	    	  <option value="{{.}}">{{.}}</option>
    	      {{end}}
          </select>
          <button type="submit">Submit</button> 
        </form>
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


      // ****************************************** js code for the activity segments section *****************************************************
      
        const bar = document.querySelector(".bar");
        const markersContainer = document.getElementById("markers");
        const datePicker = document.getElementById("datePicker");

        const today = new Date().toISOString().split("T")[0];
        datePicker.value = today;

        for (let hour = 0; hour <= 24; hour++) {
          const marker = document.createElement("div");
          marker.className = "marker";
          if (hour % 6 === 0) marker.classList.add("major");
          marker.style.left = (hour / 24) * 100 + "%";
          markersContainer.appendChild(marker);
        }

        datePicker.addEventListener("change", (e) => {
          updateView(e.target.value);
        }); 

        async function updateView(date) {
            const segments = await fetchSegments(date);
            const dayStart = getDayStart(date);
        
            renderSegments(segments, dayStart);
        }
        
        async function fetchSegments(date) {
          const url = "/segments?date=" + date;
          var res = await fetch(url);
          res = await res.json();
          segments = res.Segments;
          return segments
        }

        function getDayStart(dateStr) {
          const d = new Date(dateStr);
          d.setHours(0, 0, 0, 0);
          return Math.floor(d.getTime() / 1000);
        }

        // initial load
        updateView(today);

        function renderSegments(segments, dayStart) {
          clearSegments();
        
          segments.forEach(({ start, end, activity }) => {
            const clampedStart = Math.max(start, dayStart);
            const clampedEnd = Math.min(end, dayStart + 86400);
        
            const width = ((clampedEnd - clampedStart) / 86400) * 100;
            if (width <= 0) {
                console.log("invalid so skipping. Segment width is less than 0")
            }
        
            const left = ((clampedStart - dayStart) / 86400) * 100;
            const duration = clampedEnd - clampedStart;
        
            const segment = document.createElement("div");
            segment.className = "segment";
            segment.style.left = left + "%";
            segment.style.width = width + "%";
    
            var start_date = new Date(start * 1000);
            var start_hours = start_date.getHours();
            var start_minutes = "0" + start_date.getMinutes();
            var start_time = start_hours + ':' + start_minutes.substr(-2);

            var end_date = new Date(end * 1000);
            var end_hours = end_date.getHours();
            var end_minutes = "0" + end_date.getMinutes();
            var end_time = end_hours + ':' + end_minutes.substr(-2);

            const content = start_time + " - " + end_time + "<br>" + activity + ": " +  formatDuration(duration);
        
            segment.addEventListener("mousemove", (e) => {
              showTooltip(e, content);
            });
        
            segment.addEventListener("mouseleave", hideTooltip);
        
            bar.appendChild(segment);
          });
        }
      
        function clearSegments() {
          bar.querySelectorAll(".segment").forEach(el => el.remove());
        }

        function formatDuration(seconds) {
          const h = Math.floor(seconds / 3600);
          const m = Math.floor((seconds % 3600) / 60);
          if (h && m) return h + " hr(s)" + " & " + m + " min(s)";
          if (h) return h + " hr(s)";
          return m + " min(s)";
        }        
 
        function showTooltip(e, content) {
          segmenttooltip.innerHTML = content;
          segmenttooltip.style.opacity = 1;
          positionTooltip(e);
        }

        function hideTooltip() {
          segmenttooltip.style.opacity = 0;
        }

        function positionTooltip(e) {
          const offset = 10;
            let x = e.clientX + offset;
            let y = e.clientY + offset;
            const rect = segmenttooltip.getBoundingClientRect();
            // flip horizontally if overflowing right
            if (x + rect.width > window.innerWidth) {
              x = e.clientX - rect.width - offset;
            }      
            // flip vertically if overflowing bottom
            if (y + rect.height > window.innerHeight) {
              y = e.clientY - rect.height - offset;
            }
          segmenttooltip.style.left = x + "px";
          segmenttooltip.style.top = y + "px";
        }

    // ****************************************** js code for the heat map section *****************************************************
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
<form hx-get="/summary" hx-trigger="submit" hx-target="#activity-chart">
  <select name="year"> 
  	{{range .YearOptions}}
  	  <option value="{{.}}">{{.}}</option>
    {{end}}
  </select>
  <button type="submit">Submit</button> 
</form>
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
