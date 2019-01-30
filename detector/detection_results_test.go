package detector

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDetectionResultsAreSuccessful(t *testing.T) {
	results := NewDetectionResults()
	assert.True(t, results.Successful(), "New detection result is always expected to succeed")
	assert.False(t, results.HasFailures(), "New detection result is not expected to fail")
}

func TestCallingFailOnDetectionResultsFails(t *testing.T) {
	results := NewDetectionResults()
	results.Fail("some_filename", "Bomb", []string{})
	assert.False(t, results.Successful(), "Calling fail on a result should not make it succeed")
	assert.True(t, results.HasFailures(), "Calling fail on a result should make it fail")
}

func TestCanRecordMultipleErrorsAgainstASingleFile(t *testing.T) {
	results := NewDetectionResults()
	results.Fail("some_filename", "Bomb", []string{})
	results.Fail("some_filename", "Complete & utter failure", []string{})
	results.Fail("another_filename", "Complete & utter failure", []string{})
	assert.Len(t, results.GetFailures("some_filename").FailuresInCommits, 2, "Expected two errors against some_filename.")
	assert.Len(t, results.GetFailures("another_filename").FailuresInCommits, 1, "Expected one error against another_filename")
}

func TestResultsReportsFailures(t *testing.T) {
	results := NewDetectionResults()
	results.Fail("some_filename", "Bomb", []string{})
	results.Fail("some_filename", "Complete & utter failure", []string{})
	results.Fail("another_filename", "Complete & utter failure", []string{})

	actualErrorReport := results.ReportFileFailures("some_filename")
	fmt.Println(actualErrorReport)

	assert.Regexp(t, "some_filename", actualErrorReport[0][0], "Error report does not contain expected output")
	assert.Regexp(t, "Bomb", actualErrorReport[0][1], "Error report does not contain expected output")
	assert.Regexp(t, "some_filename", actualErrorReport[1][0], "Error report does not contain expected output")
	assert.Regexp(t, "Complete & utter failure", actualErrorReport[1][1], "Error report does not contain expected output")
}

// Presently not showing the ignored files in the log
// func TestLoggingIgnoredFilesDoesNotCauseFailure(t *testing.T) {
// 	results := NewDetectionResults()
// 	results.Ignore("some_file", "some-detector")
// 	results.Ignore("some/other_file", "some-other-detector")
// 	results.Ignore("some_file_ignored_for_multiple_things", "some-detector")
// 	results.Ignore("some_file_ignored_for_multiple_things", "some-other-detector")
// 	assert.True(t, results.Successful(), "Calling ignore should keep the result successful.")
// 	assert.True(t, results.HasIgnores(), "Calling ignore should be logged.")
// 	assert.False(t, results.HasFailures(), "Calling ignore should not cause a result to fail.")

// 	assert.Regexp(t, "some_file was ignored by .talismanrc for the following detectors: some-detector", results.Report(), "foo")
// 	assert.Regexp(t, "some/other_file was ignored by .talismanrc for the following detectors: some-other-detector", results.Report(), "foo")
// 	assert.Regexp(t, "some_file_ignored_for_multiple_things was ignored by .talismanrc for the following detectors: some-detector, some-other-detector", results.Report(), "foo")
// }

func TestTalismanRCSuggestionWhenThereAreFailures(t *testing.T) {
	results := NewDetectionResults()
	results.Fail("some_file.pem", "Bomb", []string{})

	actualErrorReport := results.Report()

	assert.Regexp(t, "fileignoreconfig:", actualErrorReport, "Error report does not contain expected output")
	assert.Regexp(t, "- filename: some_file.pem", actualErrorReport, "Error report does not contain expected output")
	assert.Regexp(t, "checksum: 87139cc4d975333b25b6275f97680604add51b84eb8f4a3b9dcbbc652e6f27ac", actualErrorReport, "Error report does not contain expected output")
	assert.Regexp(t, "ignore_detectors: \\[\\]", actualErrorReport, "Error report does not contain expected output")

}

func TestTalismanRCSuggestionWhenNoFailures(t *testing.T) {
	results := NewDetectionResults()

	actualErrorReport := results.Report()

	assert.NotRegexp(t, "fileignoreconfig:", actualErrorReport, "Error report should not contain this output")

}
