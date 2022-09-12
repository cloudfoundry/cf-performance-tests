package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/onsi/ginkgo/v2/types"
	"github.com/onsi/gomega/gmeasure"
)

type JsonReporter struct {
	testSuiteName       string
	testHeadlineName    string
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
}

func NewJsonReporter(outputFile string, testHeadlineName string, cfDeploymentVersion string, CapiVersion string, timestamp int64, testSuiteName string, ccdbVersion string) *JsonReporter {
	return &JsonReporter{
		testSuiteName:       testSuiteName,
		testHeadlineName:    testHeadlineName,
		outputFile:          outputFile,
		CfDeploymentVersion: cfDeploymentVersion,
		CapiVersion:         CapiVersion,
		CCDBVersion:         ccdbVersion,
		Timestamp:           timestamp,
		Measurements:        map[string]map[string]Measurement{},
	}
}

func GenerateReports(reporter *JsonReporter, report types.Report) {
	for _, r := range report.SpecReports {
		for _, re := range r.ReportEntries {
			// Set up experiment
			var a interface{} = re.Value.GetRawValue()
			e := a.(*gmeasure.Experiment)

			// Set up measurement
			m := Measurement{}
			m.Name = "request time"

			// Attach all results for experiment to measurement
			exp := e.Get(e.Measurements[0].Name)
			durations := exp.Durations
			var floatDurations []float64

			for _, d := range durations {
				floatDurations = append(floatDurations, d.Seconds())
			}

			m.Results = floatDurations

			// Attach experiment statistics to measurement
			expStats := e.GetStats(e.Measurements[0].Name)
			m.Smallest = expStats.DurationBundle[gmeasure.StatMin].Seconds()
			m.Largest = expStats.DurationBundle[gmeasure.StatMax].Seconds()
			m.Average = expStats.DurationBundle[gmeasure.StatMean].Seconds()
			m.StdDeviation = expStats.DurationBundle[gmeasure.StatStdDev].Seconds()

			// Attach labels to measurement
			m.SmallestLabel = "Smallest"
			m.LargestLabel = "Largest"
			m.AverageLabel = "Average"
			m.Units = "Seconds"

			// Create measurement map structure
			mp := make(map[string]Measurement)
			mp["request time"] = m

			// Add map to overall reporter structure
			reporter.Measurements[fmt.Sprintf("%s::%s", reporter.testHeadlineName, e.Name)] = mp
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
