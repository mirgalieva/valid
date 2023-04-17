package homework

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	arr := make([]string, len(v))
	for i, el := range v {
		arr[i] = el.Err.Error()
	}
	err := strings.Join(arr, "\n")
	return err
}

func Validate(v any) error {
	value := reflect.ValueOf(v)
	vType := reflect.TypeOf(v)
	if value.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	var validateErrs ValidationErrors
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		fieldValue := value.Field(i)
		fieldType := vType.Field(i)
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}
		if !fieldType.IsExported() {
			validateErrs = append(validateErrs, ValidationError{Err: ErrValidateForUnexportedFields})
			continue
		}
		fieldError := validateField(fieldValue, validateTag)
		if fieldError != nil {
			validateErrs = append(validateErrs, ValidationError{fieldError})
		}
	}

	if len(validateErrs) == 0 {
		return nil
	}
	return validateErrs
}

func validateField(value reflect.Value, validateTag string) error {
	validatorParts := strings.Split(validateTag, ",")
	validatorMap, err := parseValidator(validatorParts)
	if err != nil {
		return err
	}
	for validatorName, validatorArgs := range validatorMap {
		switch validatorName {
		case "len":
			if validatorArgs == nil {
				return ErrInvalidValidatorSyntax
			}
			expectedLength, err := strconv.Atoi(validatorArgs[0])
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			actualLength := len(value.String())
			if actualLength != expectedLength {
				return fmt.Errorf("expected length %d, actual length %d", expectedLength, actualLength)
			}
		case "in":
			if !inSlice(value, validatorArgs) {
				return fmt.Errorf("value not in %v", validatorArgs)
			}
		case "min":
			expectedMin, err := strconv.Atoi(validatorArgs[0])
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			if value.Kind() == reflect.String {
				actualLength := len(value.String())
				if actualLength < expectedMin {
					return fmt.Errorf("expected minimum length %d, actual length %d", expectedMin, actualLength)
				}
			} else if value.Kind() == reflect.Int {
				actualValue := int(value.Int())
				if actualValue < expectedMin {
					return fmt.Errorf("expected minimum value %d, actual value %d", expectedMin, actualValue)
				}
			} else {
				return fmt.Errorf("unsupported type for 'min' validator: %s", value.Type())
			}
		case "max":
			expectedMax, err := strconv.Atoi(validatorArgs[0])
			if err != nil {
				return ErrInvalidValidatorSyntax
			}
			if value.Kind() == reflect.String {
				actualLength := len(value.String())
				if actualLength > expectedMax {
					return fmt.Errorf("expected maximum length %d, actual length %d", expectedMax, actualLength)
				}
			} else if value.Kind() == reflect.Int {
				actualValue := int(value.Int())
				if actualValue > expectedMax {
					return fmt.Errorf("expected maximum value %d, actual value %d", expectedMax, actualValue)
				}
			} else {
				return fmt.Errorf("unsupported type for 'max' validator: %s", value.Type())
			}
		default:
			return fmt.Errorf("unknown validator: %s", validatorName)
		}
	}

	return nil
}

func parseValidator(validator []string) (map[string][]string, error) {
	arr := strings.Split(validator[0], ":")
	if l := len(arr); l != 2 || arr[1] == "" {
		if arr[1] == "" {
			l = 1
		}
		return nil, fmt.Errorf("wrong number of arguments in tag: %s, len: %d, want 2", arr, l)
	}
	var validatorMap = make(map[string][]string)
	validatorName := arr[0]
	if (validatorName == "min") || (validatorName == "max") {
		for i := 1; i < len(validator); i++ {
			sep := strings.Split(validator[i], ":")
			if len(sep) == 2 {
				validatorMap[sep[0]] = append(validatorMap[sep[0]], sep[1])
			}
		}
	}
	//validatorMap[validatorName] = validator[1:]
	validatorMap[validatorName] = append(validatorMap[validatorName], arr[1])
	return validatorMap, nil
}

func inSlice(value any, values []string) bool {
	for _, v := range values {
		if fmt.Sprintf("%v", value) == v {
			return true
		}
	}
	return false
}
