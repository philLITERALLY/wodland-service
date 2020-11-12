package db

import (
	"database/sql"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/philLITERALLY/wodland-service/internal/data"
)

// CreateActivity will create an Activity
func CreateActivity(db *sql.DB, activity data.ActivityInput, userID int) error {
	activityQuery := psql.
		Insert("activity").
		Columns("user_id, wod_id, date, time_taken, meps, exertion, notes").
		Values(userID, activity.WODID, activity.Date, activity.TimeTaken, activity.MEPs, activity.Exertion, activity.Notes)
	sqlActivityQuery, args, _ := activityQuery.ToSql()

	_, err := db.Exec(sqlActivityQuery, args...)
	if err != nil {
		return err
	}

	return nil
}
