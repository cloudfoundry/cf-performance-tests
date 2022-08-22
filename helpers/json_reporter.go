package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/onsi/ginkgo/v2/config"
	"github.com/onsi/ginkgo/v2/types"
)

type V2JsonReporter struct {
	Measurements        map[string]map[string]Measurement `json:"measurements"`
	outputFile          string
	CfDeploymentVersion string `json:"cfDeploymentVersion"`
	Timestamp           int64  `json:"timestamp"`
	CapiVersion         string `json:"capiVersion"`
	CCDBVersion         string `json:"ccdbVersion"`
}

type CustomReportEntry struct {
	Name    string    `json:"Name"`
	Results []float64 `json:"Results"` // TODO implement
	Number  int       `json:"N"`
	Min     float64   `json:"Min"`
	Median  float64   `json:"Median"`
	Mean    float64   `json:"Mean"`
	StdDev  float64   `json:"StdDev"`
	Max     float64   `json:"Max"`
}

type Measurement struct {
	Name          string      `json:"Name"`
	Info          interface{} `json:"Info"`
	Order         int         `json:"Order"`
	Results       []float64   `json:"Results"`
	Smallest      float64     `json:"Smallest"`
	Largest       float64     `json:"Largest"`
	Average       float64     `json:"Average"`
	StdDeviation  float64     `json:"StdDeviation"`
	SmallestLabel string      `json:"SmallestLabel"`
	LargestLabel  string      `json:"LargestLabel"`
	AverageLabel  string      `json:"AverageLabel"`
	Units         string      `json:"Units"`
	Precision     int         `json:"Precision"`
}

func NewV2JsonReporter(outputFile string, cfDeploymentVersion string, CapiVersion string, timestamp int64) *V2JsonReporter {
	return &V2JsonReporter{
		outputFile:          outputFile,
		CfDeploymentVersion: cfDeploymentVersion,
		CapiVersion:         CapiVersion,
		Timestamp:           timestamp,
		Measurements:        map[string]map[string]Measurement{},
	}
}

func V2GenerateReports(reporter *V2JsonReporter, report types.Report) {
	// For each report in SpecReports
	for _, r := range report.SpecReports {

		// For each report in ReportEntries
		for _, re := range r.ReportEntries {
			// Set up CustomReportEntry
			cre := CustomReportEntry{}
			json.Unmarshal([]byte(re.Value.AsJSON), &cre)

			// Set up measurement
			m := Measurement{}

			// Fill values
			m.Name = re.Name

			// m.Results =
			m.Smallest = cre.Min
			m.Largest = cre.Max
			m.Average = cre.Mean
			m.StdDeviation = cre.StdDev

			mp := make(map[string]Measurement)
			mp["request time"] = m

			// Add struct to V2JsonReport array of measurements
			reporter.Measurements[fmt.Sprint("domains::GET /v3/domains::as admin")] = mp // TODO fix hard coding here by using a fmt.Sprintf
		}
	}

	data, err := json.Marshal(reporter)
	if err != nil {
		fmt.Println("Failed to marshal JSON report data")
	}

	err = ioutil.WriteFile(reporter.outputFile, data, 0644)
	if err != nil {
		fmt.Println("Failed to write JSON report")
	}
}

type Reporter interface {
	SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary)
	BeforeSuiteDidRun(setupSummary *types.SetupSummary)
	SpecWillRun(specSummary *types.SpecSummary)
	SpecDidComplete(specSummary *types.SpecSummary)
	AfterSuiteDidRun(setupSummary *types.SetupSummary)
	SpecSuiteDidEnd(summary *types.SuiteSummary)
}

type JsonReporter struct {
	// only capitalised elements are exported and marshalled to json
	Measurements        map[string]map[string]*types.SpecMeasurement `json:"measurements"`
	outputFile          string
	CfDeploymentVersion string `json:"cfDeploymentVersion"`
	Timestamp           int64  `json:"timestamp"`
	CapiVersion         string `json:"capiVersion"`
}

func NewJsonReporter(outputFile string, cfDeploymentVersion string, CapiVersion string, timestamp int64) *JsonReporter {
	return &JsonReporter{
		outputFile:          outputFile,
		CfDeploymentVersion: cfDeploymentVersion,
		CapiVersion:         CapiVersion,
		Timestamp:           timestamp,
		Measurements:        map[string]map[string]*types.SpecMeasurement{},
	}
}

func (reporter *JsonReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (reporter *JsonReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *JsonReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *JsonReporter) SpecWillRun(specSummary *types.SpecSummary) {
}

func (reporter *JsonReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	specName := strings.Join(specSummary.ComponentTexts[1:], "::")
	reporter.Measurements[specName] = specSummary.Measurements
}

func (reporter *JsonReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	data, err := json.Marshal(reporter)
	if err != nil {
		fmt.Println("failed to marshal JSON report data")
	}
	err = ioutil.WriteFile(reporter.outputFile, data, 0644)
	if err != nil {
		fmt.Println("failed to write JSON report")
	}
}
