package validation

import (
	"fmt"

	"qiniupkg.com/http/httputil.v2"
)

// ---------------------------------------------------------------------------

type StringValidator interface {
	Validate(k, v string) error
}

func StringValidate(k, v string, validators ...StringValidator) error {

	for _, validator := range validators {
		err := validator.Validate(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------

type required struct {
}

func Required() required {
	return required{}
}

func (p required) Validate(k, v string) error {
	if v == "" {
		return httputil.NewError(400, k + " is required")
	}
	return nil
}

// ---------------------------------------------------------------------------

type rangeLen struct {
	Min int
	Max int
}

func RangeLen(min, max int) rangeLen {
	return rangeLen{min, max}
}

func (p rangeLen) Validate(k, v string) error {
	if len(v) < p.Min {
		return httputil.NewError(400, fmt.Sprintf("%s is too short: minimum length is %d", k, p.Min))
	}
	if p.Max != 0 && len(v) > p.Max {
		return httputil.NewError(400, fmt.Sprintf("%s is too long: maximum length is %d", k, p.Max))
	} 
	return nil
}

// ---------------------------------------------------------------------------

type minLen struct {
	Min int
}

func MinLen(min int) minLen {
	return minLen{min}
}

func (p minLen) Validate(k, v string) error {
	validator := rangeLen{p.Min, 0}
	return validator.Validate(k, v)
}

// ---------------------------------------------------------------------------

type maxLen struct {
	Max int
}

func MaxLen(max int) maxLen {
	return maxLen{max}
}

func (p maxLen) Validate(k, v string) error {
	validator := rangeLen{0, p.Max}
	return validator.Validate(k, v)
}

// ---------------------------------------------------------------------------

