package db

import (
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/philLITERALLY/wodland-service/internal/data"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// GetWOD will get and return an individual WOD
func GetWOD(db *sql.DB, wodID string, userID int) (data.WOD, error) {
	var dbWOD = data.WOD{}
	var dbActivities []data.Activity

	wodQuery := psql.
		Select("wod.*, COUNT(activity.id), MIN(activity.time_taken)").
		From("wod").
		LeftJoin("activity ON activity.wod_id = wod.id").
		Where(sq.Eq{"wod.id": wodID}).
		Where(sq.Eq{"activity.user_id": userID}).
		GroupBy("wod.id")
	sqlWODQuery, args, _ := wodQuery.ToSql()

	err := db.QueryRow(sqlWODQuery, args...).
		Scan(&dbWOD.ID, &dbWOD.Source, &dbWOD.CreationT, &dbWOD.Exercise, &dbWOD.Picture, &dbWOD.Type, &dbWOD.Attempts, &dbWOD.BestTime)
	if err != nil {
		return dbWOD, err
	}

	activityQuery := psql.
		Select("id, date, time_taken, meps, exertion, notes").
		From("activity").
		Where(sq.Eq{"wod_id": wodID}).
		Where(sq.Eq{"activity.user_id": userID})
	sqlActivityQuery, args, _ := activityQuery.ToSql()

	rows, err := db.Query(sqlActivityQuery, args...)
	if err != nil {
		return dbWOD, err
	}

	defer rows.Close()
	for rows.Next() {
		var activity data.Activity

		if err := rows.Scan(&activity.ID, &activity.Date, &activity.TimeTaken, &activity.MEPs, &activity.Exertion, &activity.Notes); err != nil {
			return dbWOD, err
		}

		dbActivities = append(dbActivities, activity)
	}

	if len(dbActivities) > 0 {
		dbWOD.Activities = &dbActivities
	}

	return dbWOD, nil
}

// GetWODs will get and return WODs
func GetWODs(db *sql.DB, filters *data.WODFilter, userID int) ([]data.WOD, error) {
	var dbWODs = []data.WOD{}

	selectQuery := psql.
		Select("wod.*, COUNT(activity.id), MIN(activity.time_taken)").
		From("wod").
		LeftJoin("activity ON activity.wod_id = wod.id").
		Where(sq.Eq{"activity.user_id": userID}).
		GroupBy("wod.id")

	selectQuery = processWODFilters(selectQuery, filters)
	selectQuery = selectQuery.Limit(10)
	sqlQuery, args, _ := selectQuery.ToSql()

	rows, err := db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var wod data.WOD

		if err := rows.Scan(&wod.ID, &wod.Source, &wod.CreationT, &wod.Exercise, &wod.Picture, &wod.Type, &wod.Attempts, &wod.BestTime); err != nil {
			return nil, err
		}

		dbWODs = append(dbWODs, wod)
	}

	return dbWODs, nil
}

func processWODFilters(baseQuery sq.SelectBuilder, filters *data.WODFilter) sq.SelectBuilder {
	baseQuery = processSourceFilter(baseQuery, filters)
	baseQuery = processWODDateFilter(baseQuery, filters)
	baseQuery = processExerciseFilter(baseQuery, filters)
	baseQuery = processPictureFilter(baseQuery, filters)
	baseQuery = processTypeFilter(baseQuery, filters)
	baseQuery = processTriedFilter(baseQuery, filters)

	return baseQuery
}

func processSourceFilter(baseQuery sq.SelectBuilder, filters *data.WODFilter) sq.SelectBuilder {
	if len(filters.Source) > 0 {
		baseQuery = baseQuery.Where(sq.Expr("LOWER(wod.source) LIKE LOWER(?)", fmt.Sprint("%", filters.Source, "%")))
	}
	return baseQuery
}

func processWODDateFilter(baseQuery sq.SelectBuilder, filters *data.WODFilter) sq.SelectBuilder {
	if !filters.StartDate.IsZero() {
		baseQuery = baseQuery.Where("wod.creation_t >= ?", float64(filters.StartDate.Unix()))
	}

	if !filters.EndDate.IsZero() {
		baseQuery = baseQuery.Where("wod.creation_t <= ?", float64(filters.EndDate.Unix()))
	}

	return baseQuery
}

func processExerciseFilter(baseQuery sq.SelectBuilder, filters *data.WODFilter) sq.SelectBuilder {
	if len(filters.Exercise) > 0 {
		clause := sq.And{}
		for _, exercise := range filters.Exercise {
			clause = append(clause, sq.Expr("LOWER(wod.wod) LIKE LOWER(?)", fmt.Sprint("%", exercise, "%")))
		}
		baseQuery = baseQuery.Where(clause)
	}
	return baseQuery
}

func processPictureFilter(baseQuery sq.SelectBuilder, filters *data.WODFilter) sq.SelectBuilder {
	if filters.Picture == nil {
		return baseQuery
	}

	if *filters.Picture == true {
		baseQuery = baseQuery.Where(sq.NotEq{"wod.picture": nil})
	}

	if *filters.Picture == false {
		baseQuery = baseQuery.Where(sq.Eq{"wod.picture": nil})
	}

	return baseQuery
}

func processTypeFilter(baseQuery sq.SelectBuilder, filters *data.WODFilter) sq.SelectBuilder {
	if len(filters.Type) > 0 {
		baseQuery = baseQuery.Where(sq.Expr("LOWER(wod.type) LIKE LOWER(?)", fmt.Sprint("%", filters.Type, "%")))
	}
	return baseQuery
}

func processTriedFilter(baseQuery sq.SelectBuilder, filters *data.WODFilter) sq.SelectBuilder {
	if filters.Tried == nil {
		return baseQuery
	}

	if *filters.Tried == true {
		baseQuery = baseQuery.Where(sq.NotEq{"activity.id": nil})
	}

	if *filters.Tried == false {
		baseQuery = baseQuery.Where(sq.Eq{"activity.id": nil})
	}

	return baseQuery
}
