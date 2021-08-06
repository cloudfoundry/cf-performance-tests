package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

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
		CapiVersion:		 CapiVersion,
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
