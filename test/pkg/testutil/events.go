package testutil

import (
	"encoding/json"
	"os"
	"slices"
	"sync"
	"text/template"
	"time"
)

type EventManager struct {
	name   string
	events []IEvent
	suites map[string]suite
	mu     sync.Mutex
	start  time.Time
}

type suite struct {
	name   string
	start  time.Time
	events []IEvent
}

func NewEventManager(name string) *EventManager {
	return &EventManager{
		name:   name,
		suites: make(map[string]suite),
		start:  time.Now(),
	}
}

func (em *EventManager) AddEvent(e IEvent) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.events = append(em.events, e)

	if s, ok := em.suites[e.GetSuite()]; !ok {
		em.suites[e.GetSuite()] = suite{
			name:   e.GetSuite(),
			start:  time.Now(),
			events: []IEvent{e},
		}
	} else {
		s.events = append(s.events, e)
		em.suites[e.GetSuite()] = s
	}
}

func (em *EventManager) WriteToFiles(intervalsFile, timelinesFile string) error {
	ivs := em.GetIntervals()

	contents, err := json.MarshalIndent(ivs, "", "    ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(intervalsFile, contents, 0644); err != nil {
		return err
	}

	data := struct {
		Title string
		Data  string
	}{
		Title: "MicroShift Test Harness - Compose phase timelines",
		Data:  string(contents),
	}

	tpl, err := template.New("timelines").Parse(timelines)
	if err != nil {
		return err
	}
	f, err := os.Create(timelinesFile)
	if err != nil {
		return err
	}
	return tpl.Execute(f, data)
}

func (em *EventManager) GetIntervals() []Interval {
	intervals := []Interval{}

	for _, suite := range em.suites {
		for _, event := range suite.events {
			intervals = append(intervals, event.GetInterval())
		}
	}

	slices.SortFunc(intervals, func(a Interval, b Interval) int {
		return int(a.Start.Sub(b.Start).Microseconds())
	})

	return intervals
}

func (em *EventManager) GetJUnit() *JUnitTestSuites {
	jsuites := &JUnitTestSuites{
		Name:      em.name,
		Time:      int(time.Since(em.start).Seconds()),
		Timestamp: em.start,
	}

	for suiteName, suite := range em.suites {
		jsuite := JUnitTestSuite{
			Name:      suiteName,
			Time:      int(time.Since(suite.start).Seconds()),
			Timestamp: suite.start,
		}

		for _, event := range suite.events {
			jtest := event.GetJUnitTestCase()
			jsuite.TestCases = append(jsuite.TestCases, jtest)
			jsuite.Tests += 1
			if jtest.Failure != nil {
				jsuite.Failures += 1
			} else if jtest.Skipped != nil {
				jsuite.Skipped += 1
			}
		}

		jsuites.TestSuites = append(jsuites.TestSuites, jsuite)
	}

	return jsuites
}

type IEvent interface {
	GetJUnitTestCase() JUnitTestCase
	GetSuite() string
	GetInterval() Interval
}

var _ IEvent = (*Event)(nil)

type Event struct {
	Name      string
	Suite     string
	ClassName string
	Start     time.Time
	End       time.Time
	SystemOut string
}

func (e *Event) GetInterval() Interval {
	return Interval{
		Name:      e.Name,
		Suite:     e.Suite,
		ClassName: e.ClassName,
		Start:     e.Start,
		End:       e.End,
		Result:    "ok",
	}
}

func (e *Event) GetSuite() string {
	return e.Suite
}

func (e *Event) GetJUnitTestCase() JUnitTestCase {
	return JUnitTestCase{
		Name:      e.Name,
		ClassName: e.ClassName,
		Time:      int(e.End.Sub(e.Start).Seconds()),
		SystemOut: e.SystemOut,
	}
}

var _ IEvent = (*SkippedEvent)(nil)

type SkippedEvent struct {
	Event
	Message string
}

func (e *SkippedEvent) GetJUnitTestCase() JUnitTestCase {
	tc := e.Event.GetJUnitTestCase()
	tc.Skipped = &JUnitSkipped{
		Message: e.Message,
	}
	return tc
}

func (e *SkippedEvent) GetInterval() Interval {
	i := e.Event.GetInterval()
	i.Result = "skipped"
	return i
}

var _ IEvent = (*FailedEvent)(nil)

type FailedEvent struct {
	Event
	Message string
	Content string
}

func (e *FailedEvent) GetJUnitTestCase() JUnitTestCase {
	tc := e.Event.GetJUnitTestCase()
	tc.Failure = &JUnitFailure{
		Message: e.Message,
		Content: e.Content,
	}
	return tc
}

func (e *FailedEvent) GetInterval() Interval {
	i := e.Event.GetInterval()
	i.Result = "failed"
	return i
}

type Interval struct {
	Name      string
	Suite     string
	ClassName string
	Start     time.Time
	End       time.Time
	Result    string
}

// adapted from https://github.com/openshift/origin/tree/master/e2echart
const timelines = `
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>{{ .Title }}</title>
    <script src="https://unpkg.com/timelines-chart"></script>
    <script src="https://d3js.org/d3-array.v1.min.js"></script>
    <script src="https://d3js.org/d3-collection.v1.min.js"></script>
    <script src="https://d3js.org/d3-color.v1.min.js"></script>
    <script src="https://d3js.org/d3-format.v1.min.js"></script>
    <script src="https://d3js.org/d3-interpolate.v1.min.js"></script>
    <script src="https://d3js.org/d3-time.v1.min.js"></script>
    <script src="https://d3js.org/d3-time-format.v2.min.js"></script>
    <script src="https://d3js.org/d3-scale.v2.min.js"></script>
    <script src="https://code.jquery.com/jquery-3.2.1.slim.min.js"
        integrity="sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN"
        crossorigin="anonymous"></script>
    <style>
        body {
            background-color: #000000;
        }
  </style>

</head>

<body>
    <div id="chart"></div>

    <script>
        var eventIntervals = {
            "items": {{ .Data }}
        }
    </script>

    <script>
        // Re-render the chart with input as a regexp. Timeout for event debouncing.
        $('#filterInput').on('input', (e) => {
            var $this = $(this);
            clearTimeout($this.data('timeout'));
            $this.data('timeout', setTimeout(() => {
                document.getElementById("chart").innerHTML = "";
                renderChart(new RegExp(e.target.value))
            }, 250));
        });

        // Prevent page refresh from pressing enter in input box
        $('#filterInput').keypress((e) => {
            if (event.which == '13') {
                event.preventDefault();
            }
        });

        function renderChart(regex) {
            var loc = window.location.href;

            var input = [];
            eventIntervals.items.forEach((interval) => {
                var fgroups = input.filter((g) => g.group == interval.Suite);
                const start = new Date(interval.Start);
                const end = new Date(interval.End);


                const durationSeconds = Math.ceil(Math.abs(end - start) / 1000);
                const seconds = durationSeconds % 60;
                const minutes = (durationSeconds - seconds) / 60;
                const d = {
                    label: interval.Name + "_" + interval.ClassName,
                    data: [
                        {
                            timeRange: [start, end],
                            val: interval.Result,
                            labelVal: "Duration: " + minutes + "m " + seconds + "s"
                        }
                    ]
                };
                if (fgroups.length === 0) {
                    input.push({
                        group: interval.Suite,
                        data: [d]
                    });
                } else if (fgroups.length === 1) {
                    var group = fgroups[0];
                    group.data.push(d);
                } else {
                    throw new Error("wrong result fgroups: " + fgroups)
                }
            });



            const el = document.querySelector('#chart');
            const myChart = TimelinesChart();
            var ordinalScale = d3.scaleOrdinal()
                .domain(['ok', 'failed', "skipped"]).
                range(['#00ff00', '#ff0000', '#bbbbbb']);

            myChart.
                data(input).
                useUtc(true).
                zQualitative(true).
                enableAnimations(false).
                leftMargin(240).
                rightMargin(1550).
                maxLineHeight(20).
                maxHeight(10000).
                zColorScale(ordinalScale)
                (el);

            // force a minimum width for smaller devices (which otherwise get an unusable display)
            setTimeout(() => { if (myChart.width() < 3100) { myChart.width(3100) } }, 1)
        }

        renderChart(null)
    </script>
</body>

</html>`
