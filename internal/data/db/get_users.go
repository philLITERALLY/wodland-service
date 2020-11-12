package db

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/philLITERALLY/wodland-service/internal/data"
)

// GetUser will get and return a user
func GetUser(db *sql.DB, user data.Login) (data.User, error) {
	var dbUser = data.User{}

	selectQuery := psql.
		Select("id, username, password, role").
		From("\"user\"").
		Where(sq.Eq{"username": user.Username}).
		Where(sq.Eq{"password": user.Password})
	sqlQuery, args, _ := selectQuery.ToSql()

	err := db.QueryRow(sqlQuery, args...).
		Scan(&dbUser.ID, &dbUser.Username, &dbUser.Password, &dbUser.Role)
	if err != nil {
		return dbUser, err
	}

	return dbUser, nil
}
