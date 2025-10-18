package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"text/template"
)

var (
	t                *template.Template
	tmplData         *templateData
	tmplDataEnvelope *templateDataEnvelope
	mu               *sync.Mutex = &sync.Mutex{}
)

func updateTemplateData(endSession bool) error {
	mu.Lock()
	defer mu.Unlock()
	if endSession {
		todaysSummary, err := todaysSummary()
		if err != nil {
			return err
		}
		tmplData.TodaysData = transformDataForTodaysSessions(todaysSummary)
	}
	currentSession, err := getCurrentSession()
	if err != nil {
		return err
	}
	tmplDataEnvelope.ActiveSession = ""
	if currentSession.Id != 0 {
		tmplDataEnvelope.ActiveSession = currentSession.Activity
	}
	tmplDataJSON, err := writeJSON()
	if err != nil {
		return fmt.Errorf("error marshalling template data to JSON")
	}
	tmplDataEnvelope.TmplDataJSON = string(tmplDataJSON)
	return nil
}

func addFlashErrorMessageToTemplateData(msg string) {
	mu.Lock()
	defer mu.Unlock()
	tmplDataEnvelope.FlashErrorMessage = msg
}

func computeTemplateData() error {
	tpl, err := template.New("home").Parse(HOME_PAGE)
	if err != nil {
		return err
	}
	t = tpl
	tmplData, tmplDataEnvelope = &templateData{}, &templateDataEnvelope{}

	// compute current year's monthly sessions
	monthlyActivitySessionsDB, err := getTimeSpentOnEachActivityMonthly()
	if err != nil {
		return err
	}
	// for _, mas := range monthlyActivitySessionsDB {
	// 	fmt.Printf("month: %d, activity: %s, duration: %.2f\n", mas.Month, mas.Activity, mas.Duration)
	// }
	monthlyActivitySessions := transformDataForCurrentYearSessions(monthlyActivitySessionsDB)
	tmplData.CurrentYearMonthlyData = monthlyActivitySessions

	// compute over the years
	overTheYearsActivitiesSessionsDB, err := getTimeSpentOnEachActivityOverTheYears()
	if err != nil {
		return err
	}
	overTheYearsActivitiesSessions := transformDataForOverAllYearsSessions(overTheYearsActivitiesSessionsDB)
	tmplData.OverTheYearsActivitySessions = overTheYearsActivitiesSessions

	// compute today's sessions
	err = updateTemplateData(true)
	if err != nil {
		return err
	}

	tmplDataJSON, err := writeJSON()
	if err != nil {
		return fmt.Errorf("error marshalling template data to JSON")
	}
	tmplDataEnvelope.TmplDataJSON = string(tmplDataJSON)
	return nil
}

func writeJSON() ([]byte, error) {
	js, err := json.MarshalIndent(tmplData, "", "\t")
	if err != nil {
		return nil, err
	}
	js = append(js, '\n')
	return js, nil
}

func renderTemplate() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := t.Execute(buf, tmplDataEnvelope)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const HOME_PAGE = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>GoTimeit</title>
  <style>
    .vl {
      border-left: 3px solid grey;
      height: 350px;
    }
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
  </style>
</head>
<body>
  <!--{{if .FlashErrorMessage}} 
    <div id="flash-message" style="background-color: #feddc7; color: #92400e; padding: 1rem; border: 1px solid #facc15; border-radius: 6px; margin: 1rem 0; position: relative;">
      <span>This is a flash message.</span>
      <button onclick="document.getElementById('flash-message').style.display='none'"
            style="position: absolute; right: 0.5rem; top: 0.5rem; background: none; border: none; font-size: 1.2rem; cursor: pointer;">
        &times;
      </button>
    </div>
  {{end}}-->
  
  <div class="card">
    <div id="user-action">
      {{if .ActiveSession}} 
        <div class="instruction">Session for the activity {{.ActiveSession}} is currently active. To start a new session click Stop first to end the current session</div>
        <form action="/sessions" method="get">
          <input type="hidden" name="action" value="end" />
          <button type="submit" style="background-color: red; width: 100%; margin-top: 9px;">
            Stop
          </button>
        </form>
      {{else}} 
        <div class="instruction">Enter your activity name below and click Start to begin a new session.</div>
        <form id="start-session-form" method="get">
          <input type="text" name="activity" placeholder="Enter activity name" required />
          <input type="hidden" name="action" value="start" />
          <button type="submit">Start Session</button>
        </form>
      {{end}}
    </div>
  </div>
  
  <div style="display: flex; margin: auto; align-items: center; justify-content: center;">
    <div id="today"></div>
    <div class="vl"></div>
    <div id="currentyear_monthly" style="margin-top: 16px;"></div>
    <div class="vl"></div>
    <div id="over_the_years" style="margin-top: 16px;"></div>
  </div>
  
  <script src="https://cdn.jsdelivr.net/npm/apexcharts"></script>
  <script type="application/json" id="activity-data">{{ .TmplDataJSON }}</script>
  <script>

    {{if not .ActiveSession}}
      document.getElementById("start-session-form").addEventListener("submit", function (event) {
        event.preventDefault();

        const form = event.target;
        const activityName = encodeURIComponent(form.activity.value);

        const url = "/sessions/" + activityName + "?action=start";
        window.location.href = url;
      });
    {{end}}  


    const jsonString = document.getElementById("activity-data").textContent;
    const tmplData = JSON.parse(jsonString)
    
    var today = {
      chart: {
        height: 350,
        width: 300,
        type: "pie",
        horizontalAlign: "center",
        animations: {
          enabled: false
        },
      },
      legend: {
          position: 'bottom',
      },
      title: {
        text: 'Time spent on activities today',
        offsetX: 25,
      },
      series: tmplData['todays_data']['series'],
      labels: tmplData['todays_data']['labels']
    };
    var today_chart = new ApexCharts(document.querySelector("#today"), today);
    today_chart.render();

    var currentyear_monthly = {
      series: tmplData['monthly_data']['series'],
        chart: {
        type: 'bar',
        height: 350,
        width: 600,
        stacked: true,
      },
      plotOptions: {
        bar: {
          horizontal: true,
          dataLabels: {
            total: {
              enabled: true,
              offsetX: 0,
              style: {
                fontSize: '13px',
                fontWeight: 900
              }
            }
          }
        },
      },
      stroke: {
        width: 1,
        colors: ['#fff']
      },
      title: {
        text: tmplData['monthly_data']['title'],
      },
      xaxis: {
        categories: ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'],
        labels: {
          formatter: function (val) {
            return val + " hrs"
          }
        }
      },
      yaxis: {
        title: {
          text: undefined
        },
      },
      tooltip: {
        y: {
          formatter: function (val) {
            return val + " hrs"
          }
        }
      },
      fill: {
        opacity: 1
      },
      legend: {
        position: 'bottom',
        horizontalAlign: 'center'
      }
    };
    var currentyear_monthly_chart = new ApexCharts(document.querySelector("#currentyear_monthly"), currentyear_monthly);
    currentyear_monthly_chart.render();

    var overall = {
      chart: {
        height: 350,
        width: 450,
        type: "line",
      },
      stroke: tmplData['overall_data']['stroke'],
      animations: {
          enabled: false
      },
      title: {
        text: 'Overall',
      },
      series: tmplData['overall_data']['series'],
      xaxis: {
        categories: tmplData['overall_data']['categories'],
      }
    };
    var overall_chart = new ApexCharts(document.querySelector("#over_the_years"), overall);
    overall_chart.render();
  </script>
</body>
`
