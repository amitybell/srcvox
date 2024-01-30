package appstate

import "fmt"

type AppError struct {
	Fatal   bool   `json:"fatal"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func AppErr(pfx string, err error) *AppError {
	if err != nil {
		if pfx != "" {
			return &AppError{Message: fmt.Sprintf("%s: %s", pfx, err)}
		}
		return &AppError{Message: err.Error()}
	}
	return nil
}
