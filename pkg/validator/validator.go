package validator

import "errors"

func Required(value, field string) error {
	if value == "" {
		return errors.New(field + " is required")
	}
	return nil
}
