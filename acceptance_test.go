package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"talisman/git_testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const awsAccessKeyIDExample string = "accessKey=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

const talismanRCDataWithIgnoreDetectorWithFilename = `
fileignoreconfig:
- filename: private.pem
  checksum: 05db785bf1e1712f69b81eeb9956bd797b956e7179ebe3cb7bb2cd9be37a24c
  ignore_detectors: [filename]
`

const talismanRCDataWithIgnoreDetectorWithFilecontent = `
fileignoreconfig:
- filename: private.pem
  checksum: 05db785bf1e1712f69b81eeb9956bd797b956e7179ebe3cb7bb2cd9be37a24c
  ignore_detectors: [filecontent]
`

const talismanRCDataWithFileNameAndCorrectChecksum = `
fileignoreconfig:
- filename: private.pem
  checksum: 1db800b79e6e9695adc451f77be974dc47bcd84d42873560d7767bfca30db8b1
  ignore_detectors: []
`

func init() {
	git_testing.Logger = logrus.WithField("Environment", "Debug")
	git_testing.Logger.Debug("Accetpance test started")
}

func TestNotHavingAnyOutgoingChangesShouldNotFail(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		assert.Equal(t, 0, runTalisman(git), "Expected run() to return 0 if no input is available on stdin. This happens when there are no outgoing changes")
	})
}

func TestAddingSimpleFileShouldExitZero(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		assert.Equal(t, 0, runTalisman(git), "Expected run() to return 0 and pass as no suspicious files are in the repo")
	})
}

func TestAddingSecretKeyShouldExitOne(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("private.pem", "secret")
		git.AddAndcommit("*", "add private key")

		assert.Equal(t, 1, runTalisman(git), "Expected run() to return 1 and fail as pem file was present in the repo")
	})
}

func TestAddingSecretKeyAsFileContentShouldExitOne(t *testing.T) {

	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("contains_keys.properties", awsAccessKeyIDExample)
		git.AddAndcommit("*", "add private key as content")

		assert.Equal(t, 1, runTalisman(git), "Expected run() to return 1 and fail as file contains some secrets")
	})
}

func TestAddingSecretKeyShouldExitZeroIfPEMFileIsIgnored(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("private.pem", "secret")
		git.CreateFileWithContents(".talismanrc", talismanRCDataWithFileNameAndCorrectChecksum)
		git.AddAndcommit("private.pem", "add private key")

		assert.Equal(t, 0, runTalisman(git), "Expected run() to return 0 and pass as pem file was ignored")
	})
}

func TestAddingSecretKeyShouldExitOneIfPEMFileIsPresentInTheGitHistory(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		_options := options{
			debug:   false,
			githook: PrePush,
			scan:    false,
		}
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("private.pem", "secret")
		git.CreateFileWithContents(".talismanrc", talismanRCDataWithFileNameAndCorrectChecksum)
		git.AddAndcommit("private.pem", "add private key")
		assert.Equal(t, 0, runTalismanWithOptions(git, _options), "Expected run() to return 0 and pass as pem file was ignored")
	})
}

func TestScanningSimpleFileShouldExitZero(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		_options := options{
			debug:   false,
			githook: PrePush,
			scan:    false,
		}
		git.SetupBaselineFiles("simple-file")
		assert.Equal(t, 0, runTalismanWithOptions(git, _options), "Expected run() to return 0 and pass as pem file was ignored")
	})
}

func TestChecksumCalculatorShouldExitOne(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		_options := options{
			debug:    false,
			githook:  PrePush,
			checksum: "*txt1",
		}
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("private.pem", "secret")
		git.CreateFileWithContents("another/private.pem", "secret")
		git.CreateFileWithContents("sample.txt", "password")
		assert.Equal(t, 1, runTalismanWithOptions(git, _options), "Expected run() to return 0 as given patterns are found and .talsimanrc is suggested")
	})
}

func TestShouldExitOneWhenSecretIsCommitted(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		_options := options{
			debug:   false,
			githook: PreCommit,
			scan:    false,
		}
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("sample.txt", "password=somepassword \n")
		git.Add("*")
		assert.Equal(t, 1, runTalismanWithOptions(git, _options), "Expected run() to return 1 as given patterns are found")
	})
}

func TestShouldExitZeroWhenNonSecretIsCommittedButFileContainsSecretPreviously(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		_options := options{
			debug:   false,
			githook: PreCommit,
			scan:    false,
		}
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("sample.txt", "password=somepassword \n")
		git.AddAndcommit("*", "Initial Commit With Secret")

		git.AppendFileContent("sample.txt", "some text \n")
		git.Add("*")
		assert.Equal(t, 0, runTalismanWithOptions(git, _options), "Expected run() to return 1 as given patterns are found")
	})
}

// Need to work on this test case as talismanrc does  not yet support comments
// func TestAddingSecretKeyShouldExitZeroIfPEMFilesAreIgnoredAndCommented(t *testing.T) {
// 	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
// 		git.SetupBaselineFiles("simple-file")
// 		git.CreateFileWithContents("private.pem", "secret")
// 		git.CreateFileWithContents(".talismanrc", talismanRCDataWithIgnoreDetector)
// 		git.AddAndcommit("*", "add private key")

// 		assert.Equal(t, 0, runTalisman(git), "Expected run() to return 0 and pass as pem file was ignored")
// 	})
// }

func TestAddingSecretKeyShouldExitOneIfTheyContainBadContentButOnlyFilenameDetectorWasIgnored(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("private.pem", awsAccessKeyIDExample)
		git.CreateFileWithContents(".talismanrc", talismanRCDataWithIgnoreDetectorWithFilename)
		git.AddAndcommit("private.pem", "add private key")

		assert.Equal(t, 1, runTalisman(git), "Expected run() to return 1 and fail as only filename was ignored")
	})
}

func TestAddingSecretKeyShouldExitOneIfFileNameIsSensitiveButOnlyFilecontentDetectorWasIgnored(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("private.pem", awsAccessKeyIDExample)
		git.CreateFileWithContents(".talismanrc", talismanRCDataWithIgnoreDetectorWithFilecontent)
		git.AddAndcommit("private.pem", "add private key")

		assert.Equal(t, 1, runTalisman(git), "Expected run() to return 1 and fail as only filename was ignored")
	})
}

func TestStagingSecretKeyShouldExitOneWhenPreCommitFlagIsSet(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("private.pem", "secret")
		git.Add("*")

		_options := options{
			debug:   false,
			githook: PreCommit,
		}

		assert.Equal(t, 1, runTalismanWithOptions(git, _options), "Expected run() to return 1 and fail as pem file was present in the repo")
	})
}

func TestPatternFindsSecretKey(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		_options := options{
			debug:   false,
			pattern: "./*.*",
		}

		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("private.pem", "secret")

		assert.Equal(t, 1, runTalismanWithOptions(git, _options), "Expected run() to return 1 and fail as pem file was present in the repo")
	})
}

func TestPatternFindsNestedSecretKey(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		_options := options{
			debug:   false,
			pattern: "./**/*.*",
		}

		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("some-dir/private.pem", "secret")

		assert.Equal(t, 1, runTalismanWithOptions(git, _options), "Expected run() to return 1 and fail as nested pem file was present in the repo")
	})
}

func TestPatternFindsSecretInNestedFile(t *testing.T) {
	withNewTmpGitRepo(func(git *git_testing.GitTesting) {
		_options := options{
			debug:   false,
			pattern: "./**/*.*",
		}

		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("some-dir/some-file.txt", awsAccessKeyIDExample)

		assert.Equal(t, 1, runTalismanWithOptions(git, _options), "Expected run() to return 1 and fail as nested pem file was present in the repo")
	})
}

func runTalisman(git *git_testing.GitTesting) int {
	_options := options{
		debug:   false,
		githook: PrePush,
	}
	return runTalismanWithOptions(git, _options)
}

func runTalismanWithOptions(git *git_testing.GitTesting, _options options) int {
	wd, _ := os.Getwd()
	os.Chdir(git.GetRoot())
	defer func() { os.Chdir(wd) }()
	return run(mockStdIn(git.EarliestCommit(), git.LatestCommit()), _options)
}

type Operation func(dirName string)

func withNewTmpDirNamed(dirName string, operation Operation) {
	path, err := ioutil.TempDir(os.TempDir(), dirName)
	if err != nil {
		panic(err)
	}
	operation(path)
}

type GitOperation func(*git_testing.GitTesting)

func withNewTmpGitRepo(doGitOperation GitOperation) {
	withNewTmpDirNamed("talisman-acceptance-test", func(gitPath string) {
		gt := git_testing.Init(gitPath)
		gt.RemoveHooks()
		doGitOperation(gt)
		os.RemoveAll(gitPath)
	})
}

func mockStdIn(oldSha string, newSha string) io.Reader {
	return strings.NewReader(fmt.Sprintf("master %s master %s\n", newSha, oldSha))
}
