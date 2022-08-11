package qqrobot

import (
	"testing"
)

// 2022/07/19 20:11 by fzls

func Test_generateMirrorGithubRawUrls(t *testing.T) {
	got := generateMirrorGithubRawUrls("https://github.com/fzls/djc_helper/raw/master/CHANGELOG.MD")
	if len(got) <= 1 {
		t.Errorf("generateMirrorGithubRawUrls failed")
	}
}

func Test_downloadNewVersionUsingPythonScript(t *testing.T) {
	pythonInterpreter := "D:\\_codes\\Python\\djc_helper_public\\.venv_dev\\Scripts\\python.exe"
	pythonScript := "D:\\_codes\\Python\\djc_helper_public\\download_latest_version.py"
	got, err := downloadNewVersionUsingPythonScript(pythonInterpreter, pythonScript)
	if err != nil {
		t.Errorf("downloadNewVersionUsingPythonScript err=%v", err)
	}
	if got == "" {
		t.Errorf("downloadNewVersionUsingPythonScript got = empty")
	}
}
