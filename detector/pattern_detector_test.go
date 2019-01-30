package detector

import (
	"talisman/git_repo"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldDetectPasswordPatterns(t *testing.T) {
	filename := "secret.txt"

	shouldPassDetectionOfSecretPattern(filename, []byte("Potential secret pattern : \"password\" : UnsafePassword"), t)
	shouldPassDetectionOfSecretPattern(filename, []byte("Potential secret pattern : <password data=123> jdghfakjkdha</password>"), t)
	shouldPassDetectionOfSecretPattern(filename, []byte("Potential secret pattern : <passphrase data=123> AasdfYlLKHKLasdKHAFKHSKmlahsdfLK</passphrase>"), t)
	shouldPassDetectionOfSecretPattern(filename, []byte("Potential secret pattern : <ConsumerKey>alksjdhfkjaklsdhflk12345adskjf</ConsumerKey>"), t)
	shouldPassDetectionOfSecretPattern(filename, []byte("Potential secret pattern : AWS key :"), t)
	shouldPassDetectionOfSecretPattern(filename, []byte(`Potential secret pattern : BEGIN RSA PRIVATE KEY-----
aghjdjadslgjagsfjlsgjalsgjaghjldasja
-----END RSA PRIVATE KEY`), t)
}

func TestShouldIgnorePasswordPatterns(t *testing.T) {
	results := NewDetectionResults()
	content := []byte("\"password\" : UnsafePassword")
	filename := "secret.txt"
	additions := []git_repo.Addition{git_repo.NewAddition(filename, content)}
	fileIgnoreConfig := FileIgnoreConfig{filename, "833b6c24c8c2c5c7e1663226dc401b29c005492dc76a1150fc0e0f07f29d4cc3", []string{"filecontent"}}
	ignores := TalismanRCIgnore{[]FileIgnoreConfig{fileIgnoreConfig}}

	NewPatternDetector().Test(additions, ignores, results)
	assert.True(t, results.Successful(), "Expected file %s to be ignored by pattern", filename)
}

func shouldPassDetectionOfSecretPattern(filename string, content []byte, t *testing.T) {
	results := NewDetectionResults()
	additions := []git_repo.Addition{git_repo.NewAddition(filename, content)}
	NewPatternDetector().Test(additions, TalismanRCIgnore{}, results)
	expected := getMapOfEmptyCommits(content)
	assert.Equal(t, expected, results.GetFailures(additions[0].Path).FailuresInCommits)
}

func getMapOfEmptyCommits(content []byte) map[string][]string {
	failuresInCommits := make(map[string][]string)
	failuresInCommits[string(content)] = nil
	return failuresInCommits
}
