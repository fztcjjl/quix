package main

import (
	"fmt"

	apperrors "github.com/fztcjjl/quix/core/errors"
	myappv1 "github.com/fztcjjl/quix/examples/proto-errors/gen/myapp/v1"
)

func main() {
	// Use the generated parameterless constructor
	err := myappv1.UserNotFound()
	fmt.Printf("Error: %+v\n", err)
	fmt.Printf("Code: %s, Message: %s, StatusCode: %d\n", err.Code, err.Message, err.StatusCode)

	// Use WithDetails variant
	detail := map[string]any{"user_id": 123}
	errWithDetails := myappv1.UserNotFoundWithDetails(detail)
	fmt.Printf("WithDetails: %+v\n", errWithDetails)

	// Use the error code constant
	fmt.Printf("Constant: %s\n", myappv1.UserErrorNotFoundCode)

	// Compare with manual Error construction
	manualErr := &apperrors.Error{
		Code:       "USER_ERROR_NOT_FOUND",
		Message:    "用户不存在",
		StatusCode: 404,
	}
	fmt.Printf("Manual: %+v\n", manualErr)
}
