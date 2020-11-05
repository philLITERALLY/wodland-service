package db

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/philLITERALLY/wodland-service/internal/data"
)

// GetActivities will get and return Activities
func GetActivities(db *sql.DB, filters *data.ActivityFilter) ([]data.Activity, error) {
	var dbActivities = []data.Activity{}

	selectQuery := psql.
		Select("activity.id, date, time_taken, meps, exertion, notes, wod.*").
		From("activity").
		Join("wod ON wod.id = activity.wod_id")

	selectQuery = processActivityFilters(selectQuery, filters)
	selectQuery = selectQuery.Limit(10)
	sqlQuery, args, _ := selectQuery.ToSql()

	rows, err := db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var activity data.Activity
		var wod data.WOD

		if err := rows.Scan(&activity.ID, &activity.Date, &activity.TimeTaken, &activity.MEPs, &activity.Exertion, &activity.Notes, &wod.ID, &wod.Source, &wod.CreationT, &wod.Exercise, &wod.Picture, &wod.Type); err != nil {
			return nil, err
		}

		activity.WOD = &wod
		dbActivities = append(dbActivities, activity)
	}

	return dbActivities, nil
}

func processActivityFilters(baseQuery sq.SelectBuilder, filters *data.ActivityFilter) sq.SelectBuilder {
	baseQuery = processWODIDFilter(baseQuery, filters)
	baseQuery = processActivityDateFilter(baseQuery, filters)

	return baseQuery
}

func processWODIDFilter(baseQuery sq.SelectBuilder, filters *data.ActivityFilter) sq.SelectBuilder {
	if filters.WODID != "" {
		baseQuery = baseQuery.Where(sq.Eq{"wod_id": filters.WODID})
	}
	return baseQuery
}

func processActivityDateFilter(baseQuery sq.SelectBuilder, filters *data.ActivityFilter) sq.SelectBuilder {
	if !filters.StartDate.IsZero() {
		baseQuery = baseQuery.Where("date >= ?", float64(filters.StartDate.Unix()))
	}

	if !filters.EndDate.IsZero() {
		baseQuery = baseQuery.Where("date <= ?", float64(filters.EndDate.Unix()))
	}

	return baseQuery
}
