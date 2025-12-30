package utils

import (
	"github.com/TharunNo1/payflow/internal/dto"
)

// Success returns a standardized 200/201 response
func Success(data interface{}, message string) dto.Response {
	return dto.Response{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

// Error returns a standardized error format
func Error(errs interface{}, message string) dto.Response {
	return dto.Response{
		Status:  "error",
		Message: message,
		Errors:  errs,
	}
}
