package slerror

import (
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/fatih/color"
)

type Severity string

const (
	SeverityFatal    Severity = "Fatal"
	SeverityNonFatal Severity = "NonFatal"
)

type ErrorCode string

type ErrorInfo struct {
	// The underlying Go error. Especially useful for testing.

}

type SLError struct {
	// A string we can reference in documentation
	Code ErrorCode
	// Time the error occurred. This is mostly used for ordering output
	Time time.Time
	// Underlying Go error. Especially useful for testing and logging
	Err error
	// Short header in the error message
	Heading string
	// Detailed contents of the error message
	DetailedMessage string
	// Fatal or nonfatal. We can make sure the program never proceeds if it's fatal.
	Severity Severity
}

func (slerr SLError) FmtError() string {
	return fmt.Sprintf("Error %s: %s. At: %v. %s: %s", slerr.Code, slerr.Err.Error(), slerr.Time, slerr.Heading, slerr.DetailedMessage)
}

func (slerr SLError) Error() string {
	return slerr.Error()
}

func PrintErrors(errors []error) {
	sort.Slice(errors, func(i, j int) bool {
		err1, ok1 := errors[i].(*SLError)
		err2, ok2 := errors[j].(*SLError)

		if ok1 && ok2 {
			return err1.Time.Before(err2.Time)
		}
		// Default to false if not both SLError
		return false
	})

	for _, err := range errors {
		slErr, ok := err.(*SLError)
		if !ok {
			continue
		}

		if !color.NoColor {
			color.New(color.Bold).Printf("%s: %s\n", slErr.Heading, slErr.DetailedMessage)
		} else {
			fmt.Printf("%s: %s\n", slErr.Heading, slErr.DetailedMessage)
		}
	}
}

func LogError(log *logrus.Logger, err error) {
	slErr, ok := err.(*SLError)
	if !ok {
		log.Error(err)
		return
	}

	log.WithFields(logrus.Fields{
		"code":             slErr.Code,
		"error":            slErr.Err.Error(),
		"heading":          slErr.Heading,
		"detailed_message": slErr.DetailedMessage,
		"severity":         slErr.Severity,
	}).Error(slErr.Err.Error())
}

func LogAndReturn(log *logrus.Logger, err error) error {
	LogError(log, err)
	return err
}
