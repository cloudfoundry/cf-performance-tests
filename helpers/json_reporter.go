package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/onsi/ginkgo/v2/config"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/onsi/gomega/gmeasure"
)

type V2JsonReporter struct {
	testSuiteName       string
	Measurements        map[string]map[string]Measurement `json:"measurements"`
	outputFile          string
	CfDeploymentVersion string `json:"cfDeploymentVersion"`
	Timestamp           int64  `json:"timestamp"`
	CapiVersion         string `json:"capiVersion"`
	CCDBVersion         string `json:"ccdbVersion"`
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

func NewV2JsonReporter(outputFile string, cfDeploymentVersion string, CapiVersion string, timestamp int64, testSuiteName string) *V2JsonReporter {
	return &V2JsonReporter{
		testSuiteName:       testSuiteName,
		outputFile:          outputFile,
		CfDeploymentVersion: cfDeploymentVersion,
		CapiVersion:         CapiVersion,
		Timestamp:           timestamp,
		Measurements:        map[string]map[string]Measurement{},
	}
}

func V2GenerateReports(reporter *V2JsonReporter, report types.Report) {
	for _, r := range report.SpecReports {
		for _, re := range r.ReportEntries {
			// Set up experiment
			var a interface{} = re.Value.GetRawValue()
			e := a.(*gmeasure.Experiment)

			// Set up measurement
			m := Measurement{}
			m.Name = "request time"

			// TODO Attach all results for experiment to measurement
			exp := e.Get(e.Measurements[0].Name)
			m.Results = exp.Values

			// Attach experiment statistics to measurement
			expStats := e.GetStats(e.Measurements[0].Name)
			m.Smallest = float64(expStats.DurationBundle[gmeasure.StatMin])
			m.Largest = float64(expStats.DurationBundle[gmeasure.StatMax])
			m.Average = float64(expStats.DurationBundle[gmeasure.StatMean])
			m.StdDeviation = float64(expStats.DurationBundle[gmeasure.StatStdDev])

			// Attach labels to measurement
			m.SmallestLabel = "Smallest"
			m.LargestLabel = "Largest"
			m.AverageLabel = "Average"

			// Create measurement map structure
			mp := make(map[string]Measurement)
			mp["request time"] = m

			// Add map to overall reporter structure
			reporter.Measurements[fmt.Sprintf("%s::%s", reporter.testSuiteName, e.Name)] = mp
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
