package db

import (
	"database/sql"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/philLITERALLY/wodland-service/internal/data"
)

// CreateWOD will create a WOD (and add an attempt if supplied)
func CreateWOD(db *sql.DB, WOD data.CreateWOD, userID int) error {
	wodQuery := psql.
		Insert("wod").
		Columns("source, creation_t, wod, picture, type, created_by").
		Values(WOD.Source, WOD.CreationT, WOD.Exercise, WOD.Picture, WOD.Type, userID).
		Suffix("RETURNING \"id\"")
	sqlWODQuery, wodArgs, _ := wodQuery.ToSql()

	var wodID int
	wodErr := db.QueryRow(sqlWODQuery, wodArgs...).Scan(&wodID)
	if wodErr != nil {
		return wodErr
	}

	if WOD.ActivityInput != nil {
		activity := WOD.ActivityInput
		activity.WODID = &wodID

		activityErr := CreateActivity(db, *activity, userID)
		if activityErr != nil {
			return activityErr
		}
	}

	return nil
}
