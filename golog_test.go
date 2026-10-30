package golog_test

import (
	"testing"

	"github.com/kopytkg/golog"
)

func TestLogFile(t *testing.T) {
	err := golog.EnableLogFile("a.log")
	if err != nil {
		t.Fatalf("failed to enable log file: %v", err)
	}
	golog.CLILOG = golog.DISABLED
	golog.LOGLEVEL = golog.DEBUG

	golog.Error(15)
	golog.Debug("asd")
	golog.Success("asdasd")
	golog.APPWarn("asd123")
	golog.SYSWarn("asdasd123")

	golog.Errorf("%d", 15)
	golog.Debugf("asd %s", "now")
	golog.Successf("asdasd %s", "now")
	golog.APPWarnf("asd123 %s", "now")
	golog.SYSWarnf("asdasd123 %s", "now")

}

func TestCLILog(t *testing.T) {
	golog.DisableLogFile()
	golog.CLILOG = golog.ENABLED
	golog.LOGLEVEL = golog.DEBUG

	golog.Error(15)
	golog.Debug("asd")
	golog.Success("asdasd")
	golog.APPWarn("asd123")
	golog.SYSWarn("asdasd123")

	golog.Errorf("%d", 15)
	golog.Debugf("asd %s", "now")
	golog.Successf("asdasd %s", "now")
	golog.APPWarnf("asd123 %s", "now")
	golog.SYSWarnf("asdasd123 %s", "now")
}
