package data

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

// Login is the data required for a user to login
type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// User is the data object for a User
type User struct {
	ID       int    `json:"userID"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// WODInput is the data required to create a WOD
type WODInput struct {
	Source    *string `json:"source"`
	CreationT int64   `json:"creationT"`
	Exercise  *string `json:"exercise"`
	Picture   *string `json:"picture"`
	Type      string  `json:"type"`
}

// WOD is the data object returned from the WODs endpoint
type WOD struct {
	ID int `json:"id"`
	WODInput
	Attempts   *int        `json:"attempts"`
	BestTime   *int        `json:"bestTime,omitempty"`
	Activities *[]Activity `json:"activities,omitempty"`
}

// ActivityInput is the data required to create an Activity
type ActivityInput struct {
	Date      int64   `json:"date"`
	WODID     *int    `json:"wodID,omitempty"`
	TimeTaken int64   `json:"timeTaken"`
	MEPs      *int64  `json:"meps,omitempty"`
	Exertion  *int64  `json:"exertion,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

// Activity is the data object returned for each activity
type Activity struct {
	ID int64 `json:"id"`
	ActivityInput
	WOD *WOD `json:"wod,omitempty"`
}

// CreateWOD is the data object required to add a WOD
type CreateWOD struct {
	WODInput
	*ActivityInput
}

// WODFilter is used to model filterable aspects for WODs
type WODFilter struct {
	Source    string    `json:"source"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Exercise  []string  `json:"exercise"`
	Picture   *bool     `json:"picture"`
	Type      string    `json:"type"`
	Tried     *bool     `json:"tried"`
}

// ActivityFilter is used to model filterable aspects for Activities
type ActivityFilter struct {
	WODID     string    `json:"wodID"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// WODFilters will get and return any filters applied to the WODs endpoint
func WODFilters(c *gin.Context) (filters *WODFilter, err error) {
	filters = &WODFilter{}

	if err = GetFilters(c, filters); err != nil {
		return
	}

	return
}

// ActivityFilters will get and return any filters applied to the Activities endpoint
func ActivityFilters(c *gin.Context) (filters *ActivityFilter, err error) {
	filters = &ActivityFilter{}

	if err = GetFilters(c, filters); err != nil {
		return
	}

	return
}

// GetFilters extracts filter parameters from the context
func GetFilters(c *gin.Context, filter interface{}) error {

	v := reflect.ValueOf(filter)

	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("should be called with a pointer to struct")
	}

	if c == nil {
		return errors.New("should be called with non nil context")
	}

	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := v.Type().Field(i)
		paramName := structField.Tag.Get("json")
		if paramName != "" {
			paramName = strings.Split(paramName, ",")[0]
		} else {
			runes := bytes.Runes([]byte(structField.Name))
			runes[0] = unicode.ToLower(runes[0])
			paramName = string(runes)
		}

		switch f.Kind() {
		case reflect.String:
			param := c.Query(paramName)
			if len(param) > 0 {
				f.Set(reflect.ValueOf(param))
			}
		case reflect.Array, reflect.Slice:
			switch f.Type().Elem().Kind() {
			case reflect.Int64:
				values := []int64{}
				arrayParam := strings.Split(c.Query(paramName), ",")
				if len(arrayParam) > 0 {
					for _, param := range arrayParam {
						if len(param) > 0 {
							params := strings.Split(param, ",")
							for _, p := range params {
								value, err := strconv.ParseInt(p, 10, 64)
								if err != nil {
									return fmt.Errorf("Invalid value for integer parameter '%s': '%s'", paramName, p)
								}
								values = append(values, value)
							}
						}
					}
				}
				f.Set(reflect.ValueOf(values))
			case reflect.Int:
				values := []int{}
				arrayParam := strings.Split(c.Query(paramName), ",")
				if len(arrayParam) > 0 {
					for _, param := range arrayParam {
						if len(param) > 0 {
							params := strings.Split(param, ",")
							for _, p := range params {
								value, err := strconv.ParseInt(p, 10, 0)
								if err != nil {
									return fmt.Errorf("Invalid value for integer parameter '%s': '%s'", paramName, p)
								}
								values = append(values, int(value))
							}
						}
					}
				}
				f.Set(reflect.ValueOf(values))
			case reflect.String:
				values := []string{}
				arrayParam := strings.Split(c.Query(paramName), ",")
				if len(arrayParam) > 0 {
					for _, param := range arrayParam {
						if len(param) > 0 {
							params := strings.Split(param, ",")
							for _, p := range params {
								if len(p) > 0 {
									unescapedValue, err := url.QueryUnescape(p)
									if err != nil {
										return fmt.Errorf("Invalid encoding for string parameter '%s': '%s'", paramName, p)
									}
									values = append(values, unescapedValue)
								}
							}
						}
					}
				}
				f.Set(reflect.ValueOf(values))
			}
		case reflect.Ptr:
			param := c.Query(paramName)
			if param != "" {
				switch f.Type().Elem().Kind() {
				case reflect.Bool:
					v, err := strconv.ParseBool(param)
					if err != nil {
						return fmt.Errorf("Invalid value for boolean parameter '%s': '%s'", paramName, param)
					}
					f.Set(reflect.ValueOf(&v))
				case reflect.Int:
					v, err := strconv.Atoi(param)
					if err != nil {
						return fmt.Errorf("Invalid value for integer parameter '%s': '%s'", paramName, param)
					}
					f.Set(reflect.ValueOf(&v))
				case reflect.Int64:
					v, err := strconv.ParseInt(param, 10, 64)
					if err != nil {
						return fmt.Errorf("Invalid value for integer parameter '%s': '%s'", paramName, param)
					}
					f.Set(reflect.ValueOf(&v))
				}
			}
		case reflect.Struct:
			switch f.Type() {
			case reflect.TypeOf(time.Time{}):
				param := c.Query(paramName)
				if param != "" {
					result, err := ParseDateTimeString(param)
					if err != nil || result.IsZero() {
						return fmt.Errorf("Invalid value for date format parameter '%s': '%s'", paramName, param)
					}
					f.Set(reflect.ValueOf(result))
				}
			default:
				errMsg := GetFilters(c, f.Addr().Interface())
				if errMsg != nil {
					return errMsg
				}
			}
		case reflect.Bool:
			param := c.Query(paramName)
			if len(param) > 0 {
				result, err := strconv.ParseBool(param)
				if err != nil {
					return fmt.Errorf("Invalid value for bool parameter '%s': '%s'", paramName, param)
				}
				f.Set(reflect.ValueOf(result))
			}
		default:
			return fmt.Errorf("unsupported filter type %s for param %v", f.Kind(), v.Type().Field(i).Name)
		}
	}

	return nil
}

// ParseDateTimeString parses a date string in RFC3339 format
// (a more constrained subset of ISO8601) and returns the time.Time representation of it
func ParseDateTimeString(date string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, date)

	return t, err
}
