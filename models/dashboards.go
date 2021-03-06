package models

import (
	"fmt"
	"sort"

	"github.com/kiali/kiali/kubernetes/monitoringdashboards/v1alpha1"
	"github.com/kiali/kiali/prometheus"
)

// DashboardQuery holds query parameters for a dashboard query
type DashboardQuery struct {
	prometheus.RangeQuery
	Namespace         string
	LabelsFilters     map[string]string
	AdditionalLabels  []Aggregation
	RawDataAggregator string
}

// FillDefaults fills the struct with default parameters
func (q *DashboardQuery) FillDefaults() {
	q.RangeQuery.FillDefaults()
	q.RawDataAggregator = "sum"
}

// MonitoringDashboard is the model representing custom monitoring dashboard, transformed from MonitoringDashboard k8s resource
type MonitoringDashboard struct {
	Title         string         `json:"title"`
	Charts        []Chart        `json:"charts"`
	Aggregations  []Aggregation  `json:"aggregations"`
	ExternalLinks []ExternalLink `json:"externalLinks"`
}

// Chart is the model representing a custom chart, transformed from charts in MonitoringDashboard k8s resource
type Chart struct {
	Name           string   `json:"name"`
	Unit           string   `json:"unit"`
	Spans          int      `json:"spans"`
	StartCollapsed bool     `json:"startCollapsed"`
	ChartType      *string  `json:"chartType,omitempty"`
	Min            *int     `json:"min,omitempty"`
	Max            *int     `json:"max,omitempty"`
	Metrics        []Metric `json:"metrics"`
	XAxis          *string  `json:"xAxis"`
	Error          string   `json:"error"`
}

// ConvertChart converts a k8s chart (from MonitoringDashboard k8s resource) into this models chart
func ConvertChart(from v1alpha1.MonitoringDashboardChart) Chart {
	return Chart{
		Name:           from.Name,
		Unit:           from.Unit,
		Spans:          from.Spans,
		StartCollapsed: from.StartCollapsed,
		ChartType:      from.ChartType,
		Min:            from.Min,
		Max:            from.Max,
		Metrics:        []Metric{},
		XAxis:          from.XAxis,
	}
}

// Aggregation is the model representing label's allowed aggregation, transformed from aggregation in MonitoringDashboard k8s resource
type Aggregation struct {
	Label           string `json:"label"`
	DisplayName     string `json:"displayName"`
	SingleSelection bool   `json:"singleSelection"`
}

// ConvertAggregations converts a k8s aggregations (from MonitoringDashboard k8s resource) into this models aggregations
// Results are sorted by DisplayName
func ConvertAggregations(from v1alpha1.MonitoringDashboardSpec) []Aggregation {
	uniqueAggs := make(map[string]Aggregation)
	for _, item := range from.Items {
		for _, agg := range item.Chart.Aggregations {
			uniqueAggs[agg.DisplayName] = Aggregation{Label: agg.Label, DisplayName: agg.DisplayName, SingleSelection: agg.SingleSelection}
		}
	}
	aggs := []Aggregation{}
	for _, agg := range uniqueAggs {
		aggs = append(aggs, agg)
	}
	sort.Slice(aggs, func(i, j int) bool {
		return aggs[i].DisplayName < aggs[j].DisplayName
	})
	return aggs
}

// ExternalLink provides links to external dashboards (e.g. to Grafana)
type ExternalLink struct {
	URL       string                                            `json:"url"`
	Name      string                                            `json:"name"`
	Variables v1alpha1.MonitoringDashboardExternalLinkVariables `json:"variables"`
}

// Runtime holds the runtime title and associated dashboard template(s)
type Runtime struct {
	Name          string         `json:"name"`
	DashboardRefs []DashboardRef `json:"dashboardRefs"`
}

// DashboardRef holds template name and title for a custom dashboard
type DashboardRef struct {
	Template string `json:"template"`
	Title    string `json:"title"`
}

func buildIstioAggregations(local, remote string) []Aggregation {
	aggs := []Aggregation{
		{
			Label:       fmt.Sprintf("%s_canonical_revision", local),
			DisplayName: "Local version",
		},
	}
	if remote == "destination" {
		aggs = append(aggs, Aggregation{
			Label:       "destination_service_name",
			DisplayName: "Remote service",
		})
	}
	aggs = append(aggs, []Aggregation{
		{
			Label:       fmt.Sprintf("%s_canonical_service", remote),
			DisplayName: "Remote app",
		},
		{
			Label:       fmt.Sprintf("%s_canonical_revision", remote),
			DisplayName: "Remote version",
		},
		{
			Label:       "response_code",
			DisplayName: "Response code",
		},
		{
			Label:       "grpc_response_status",
			DisplayName: "GRPC status",
		},
		{
			Label:       "response_flags",
			DisplayName: "Response flags",
		},
	}...)
	return aggs
}

// PrepareIstioDashboard prepares the Istio dashboard title and aggregations dynamically for input values
func PrepareIstioDashboard(direction, local, remote string) MonitoringDashboard {
	return MonitoringDashboard{
		Title:        fmt.Sprintf("%s Metrics", direction),
		Aggregations: buildIstioAggregations(local, remote),
		Charts:       []Chart{},
	}
}
