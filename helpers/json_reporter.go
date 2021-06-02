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
	measurements map[string]map[string]*types.SpecMeasurement
	outputFile   string
}

func NewJsonReporter(outputFile string) *JsonReporter {
	return &JsonReporter{
		outputFile:   outputFile,
		measurements: map[string]map[string]*types.SpecMeasurement{},
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
	reporter.measurements[specName] = specSummary.Measurements
}

func (reporter *JsonReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	data, err := json.Marshal(reporter.measurements)
	if err != nil {
		fmt.Println("failed to marshal JSON report data")
	}
	err = ioutil.WriteFile(reporter.outputFile, data, 0644)
	if err != nil {
		fmt.Println("failed to write JSON report")
	}
}
