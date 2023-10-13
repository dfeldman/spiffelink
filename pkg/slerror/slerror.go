package slerror

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/fatih/color"
)

// This module defines SLError, which is a highly detailed error message struct used throughout SL.
// This allows having a detailed response to any error that could occur, with instructions for remediation
// that make sense for a typical user.

// Additionally, there is an SLErrorList struct. This allows returning multiple SLErrors in situations where
// that makes sense. For example, in parsing config files, it's ideal to return a list of all errors encountered
// rather than return one and then halt. This is different from wrapped errors since these errors are not
// necessarily related.

// The detailed error messages are defined in errors.go to keep large blocks of text out of the rest of SL.

// Both SLErrorList and SLError implement the Error interface.

type Severity string

// These really should be enum and not string type
const (
	SeverityFatal    Severity = "Fatal"
	SeverityNonFatal Severity = "NonFatal"
)

type ErrorCode string

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

type SLErrorList struct {
	Errors []SLError
}

func (slerr SLError) FmtError() string {
	return fmt.Sprintf("Error %s: %s. At: %v. %s: %s", slerr.Code, slerr.Err.Error(), slerr.Time, slerr.Heading, slerr.DetailedMessage)
}

func (slerr SLError) Error() string {
	return slerr.FmtError()
}

func (slerrs SLErrorList) FmtError() string {
	var errors []string
	for _, err := range slerrs.Errors {
		errors = append(errors, err.FmtError())
	}
	return fmt.Sprintf("Multiple Errors: \n%s", join(errors, "\n"))
}

func (slerrs SLErrorList) Empty() bool {
	return len(slerrs.Errors) == 0
}

func (slerrs SLErrorList) Error() string {
	return slerrs.FmtError()
}

func New(str string) SLError {
	return SLError{
		Code: "ERROR_UNKNOWN",
		Err:  errors.New(str),
	}
}

func join(strs []string, sep string) string {
	var result string
	for i, str := range strs {
		result += str
		if i != len(strs)-1 {
			result += sep
		}
	}
	return result
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

func LogAndReturn(log *logrus.Logger, err SLError) SLError {
	LogError(log, err)
	return err
}
