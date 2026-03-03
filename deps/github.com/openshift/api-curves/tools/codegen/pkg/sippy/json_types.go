package sippy

import "fmt"

type SippyQueryStruct struct {
	Items        []SippyQueryItem `json:"items"`
	LinkOperator string           `json:"linkOperator"`
}

type SippyQueryItem struct {
	ColumnField   string `json:"columnField"`
	Not           bool   `json:"not"`
	OperatorValue string `json:"operatorValue"`
	Value         string `json:"value"`
}

type SippyTestInfo struct {
	Id                        int         `json:"id"`
	Name                      string      `json:"name"`
	SuiteName                 string      `json:"suite_name"`
	Variants                  interface{} `json:"variants"`
	JiraComponent             string      `json:"jira_component"`
	JiraComponentId           int         `json:"jira_component_id"`
	CurrentSuccesses          int         `json:"current_successes"`
	CurrentFailures           int         `json:"current_failures"`
	CurrentFlakes             int         `json:"current_flakes"`
	CurrentPassPercentage     float64     `json:"current_pass_percentage"`
	CurrentFailurePercentage  float64     `json:"current_failure_percentage"`
	CurrentFlakePercentage    float64     `json:"current_flake_percentage"`
	CurrentWorkingPercentage  float64     `json:"current_working_percentage"`
	CurrentRuns               int         `json:"current_runs"`
	PreviousSuccesses         int         `json:"previous_successes"`
	PreviousFailures          int         `json:"previous_failures"`
	PreviousFlakes            int         `json:"previous_flakes"`
	PreviousPassPercentage    float64     `json:"previous_pass_percentage"`
	PreviousFailurePercentage float64     `json:"previous_failure_percentage"`
	PreviousFlakePercentage   float64     `json:"previous_flake_percentage"`
	PreviousWorkingPercentage float64     `json:"previous_working_percentage"`
	PreviousRuns              int         `json:"previous_runs"`
	NetFailureImprovement     float64     `json:"net_failure_improvement"`
	NetFlakeImprovement       float64     `json:"net_flake_improvement"`
	NetWorkingImprovement     float64     `json:"net_working_improvement"`
	NetImprovement            float64     `json:"net_improvement"`
	Watchlist                 bool        `json:"watchlist"`
	Tags                      interface{} `json:"tags"`
	OpenBugs                  int         `json:"open_bugs"`
}

func QueriesFor(cloud, architecture, topology, networkStack, testPattern string) []*SippyQueryStruct {
	queries := []*SippyQueryStruct{
		{
			Items: []SippyQueryItem{
				{
					ColumnField:   "variants",
					Not:           false,
					OperatorValue: "contains",
					Value:         fmt.Sprintf("Platform:%s", cloud),
				},
				{
					ColumnField:   "variants",
					Not:           false,
					OperatorValue: "contains",
					Value:         fmt.Sprintf("Architecture:%s", architecture),
				},
				{
					ColumnField:   "variants",
					Not:           false,
					OperatorValue: "contains",
					Value:         fmt.Sprintf("Topology:%s", topology),
				},
				{
					ColumnField:   "name",
					Not:           false,
					OperatorValue: "contains",
					Value:         testPattern,
				},
			},
			LinkOperator: "and",
		},
	}

	if networkStack != "" {
		queries[0].Items = append(queries[0].Items,
			SippyQueryItem{
				ColumnField:   "variants",
				Not:           false,
				OperatorValue: "contains",
				Value:         fmt.Sprintf("NetworkStack:%s", networkStack),
			},
		)

	}

	return queries
}
