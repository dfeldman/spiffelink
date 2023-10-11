package slerror

import (
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestSLError_Error(t *testing.T) {
	slerr := SLError{
		Code:            "E001",
		Time:            time.Now(),
		Err:             errors.New("test error"),
		Heading:         "Test Heading",
		DetailedMessage: "This is a test error",
		Severity:        SeverityFatal,
	}

	expected := slerr.FmtError()
	if slerr.Error() != expected {
		t.Fatalf("Expected %s but got %s", expected, slerr.Error())
	}
}

func TestSLErrorList_Error(t *testing.T) {
	slerrs := SLErrorList{Errors: []SLError{
		{
			Code: "E001", Time: time.Now(),
			Err:             errors.New("test error 1"),
			Heading:         "Test Heading 1",
			DetailedMessage: "This is a test error 1",
			Severity:        SeverityFatal,
		},
		{
			Code:            "E002",
			Time:            time.Now(),
			Err:             errors.New("test error 2"),
			Heading:         "Test Heading 2",
			DetailedMessage: "This is a test error 2",
			Severity:        SeverityNonFatal,
		},
	},
	}

	expected := slerrs.FmtError()
	if slerrs.Error() != expected {
		t.Fatalf("Expected %s but got %s", expected, slerrs.Error())
	}
}

func TestLogError(t *testing.T) {
	// Using test logger from logrus
	logger, hook := test.NewNullLogger()

	slerr := SLError{
		Code:            "E001",
		Time:            time.Now(),
		Err:             errors.New("test error"),
		Heading:         "Test Heading",
		DetailedMessage: "This is a test error",
		Severity:        SeverityFatal,
	}

	LogError(logger, &slerr)

	if len(hook.Entries) != 1 {
		t.Fatalf("Expected 1 log entry but got %d", len(hook.Entries))
	}

	if hook.LastEntry().Data["code"] != slerr.Code {
		t.Fatalf("Expected log entry code to be %s but got %s", slerr.Code, hook.LastEntry().Data["code"])
	}
}
