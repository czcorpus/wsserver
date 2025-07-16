package core

import "fmt"

type ErrorType string

const (
	ErrorTypeNotFound         ErrorType = "NOT_FOUND"
	ErrorTypeInternalError    ErrorType = "INTERNAL_ERROR"
	ErrorTypeInvalidArguments ErrorType = "INVALID_ARGUMENTS"
)

type AppError struct {
	Message string
	Type    ErrorType
	Cause   error
}

func (err AppError) Error() string {
	if err.Cause != nil {
		return fmt.Sprintf("%s: %s", err.Message, err.Cause.Error())
	}
	return err.Message
}

func (err AppError) IsZero() bool {
	return err.Message == ""
}

func NewAppError(message string, etype ErrorType, cause error) AppError {
	return AppError{
		Message: message,
		Type:    etype,
		Cause:   cause,
	}
}
