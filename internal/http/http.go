package http

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/philLITERALLY/wodland-service/internal/data"
	"github.com/philLITERALLY/wodland-service/internal/data/db"
)

var usernameKey = "username"

// GetUserID will return ID of logged in User
func GetUserID(c *gin.Context) (int, error) {
	user, exists := c.Get(usernameKey)
	if !exists {
		return 0, errors.New("Error fetching logged in user")
	}

	return user.(*data.User).ID, nil
}

// GetWOD will get and return an individual WOD
func GetWOD(dataSource *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			fmt.Printf("%+v", err)
			c.JSON(http.StatusBadRequest, err)
		}

		wodID := c.Param("wodID")

		wodResult, err := db.GetWOD(dataSource, wodID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error reading wod: %q", err))
			return
		}

		c.JSON(http.StatusOK, wodResult)
	}
}

// GetWODs will get and return WODs
func GetWODs(dataSource *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			fmt.Printf("%+v", err)
			c.JSON(http.StatusBadRequest, err)
		}

		filters, err := data.WODFilters(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Error reading filters: %q", err))
			return
		}

		wodResult, err := db.GetWODs(dataSource, filters, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error reading wods: %q", err))
			return
		}

		c.JSON(http.StatusOK, wodResult)
	}
}

// AddWOD will create a WOD (and add an attempt if supplied)
func AddWOD(dataSource *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		wodInput := data.CreateWOD{}

		userID, err := GetUserID(c)
		if err != nil {
			fmt.Printf("%+v", err)
			c.JSON(http.StatusBadRequest, err)
		}

		err = c.Bind(&wodInput)
		if err != nil {
			fmt.Printf("error binding request input: %+v", err)
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Error with WOD details: %q", err))
		}

		if wodInput.Type == "" {
			c.JSON(http.StatusBadRequest, "Please provide a Type (e.g. WOD, Girls, Hero)")
			return
		}

		if wodInput.ActivityInput != nil {
			if wodInput.Date == 0 {
				c.JSON(http.StatusBadRequest, "Please provide a date for activity")
				return
			} else if wodInput.TimeTaken == 0 {
				c.JSON(http.StatusBadRequest, "Please provide a time taken for activity")
				return
			}
		}

		err = db.CreateWOD(dataSource, wodInput, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error creating WOD: %q", err))
			return
		}

		if wodInput.ActivityInput == nil {
			c.JSON(http.StatusOK, "Added a WOD")
		} else {
			c.JSON(http.StatusOK, "Added a WOD and Activity")
		}
	}
}

// GetActivities will get and return Activities
func GetActivities(dataSource *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			fmt.Printf("%+v", err)
			c.JSON(http.StatusBadRequest, err)
		}

		filters, err := data.ActivityFilters(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Error reading filters: %q", err))
			return
		}

		activityResult, err := db.GetActivities(dataSource, filters, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error reading activities: %q", err))
			return
		}

		c.JSON(http.StatusOK, activityResult)
	}
}

// AddActivity will create an Activity
func AddActivity(dataSource *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityInput := data.ActivityInput{}

		userID, err := GetUserID(c)
		if err != nil {
			fmt.Printf("%+v", err)
			c.JSON(http.StatusBadRequest, err)
		}

		err = c.Bind(&activityInput)
		if err != nil {
			fmt.Printf("error binding request input: %+v", err)
			c.JSON(http.StatusBadRequest, fmt.Sprintf("Error with Activity details: %q", err))
		}

		if activityInput.WODID == nil {
			c.JSON(http.StatusBadRequest, "Please provide a WOD ID")
			return
		} else if activityInput.Date == 0 {
			c.JSON(http.StatusBadRequest, "Please provide a date")
			return
		} else if activityInput.TimeTaken == 0 {
			c.JSON(http.StatusBadRequest, "Please provide a time taken")
			return
		}

		err = db.CreateActivity(dataSource, activityInput, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error creating activity: %q", err))
			return
		}

		c.JSON(http.StatusCreated, "Added an Activity")
	}
}
