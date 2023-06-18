package model

import (
	"context"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/traPtitech/go-traq"
	"golang.org/x/exp/slog"
)

var (
	db *sqlx.DB
)

func SetUp() error {
	_db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOSTNAME"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE")))
	if err != nil {
		slog.Info("Cannot Connect to Database: %s", err)
	}
	db = _db
	slog.Info("Connected to Database")
	err = initUsersTable()
	if err != nil {
		return err
	}
	return nil
}

var ACCESS_TOKEN = os.Getenv("BOT_ACCESS_TOKEN")

func initUsersTable() error {
	if ACCESS_TOKEN == "" {
		slog.Info("Skip initUsersTable")
		return nil
	}

	client := traq.NewAPIClient(traq.NewConfiguration())
	auth := context.WithValue(context.Background(), traq.ContextAccessToken, ACCESS_TOKEN)

	result, _, err := client.UserApi.GetUsers(auth).Execute()
	if err != nil {
		return err
	}
	for _, user := range result {
		// TODO: 50個ごとにバルクインサートするように変更する
		_, err = db.Exec("INSERT INTO users (traq_uuid, trap_id, is_bot) VALUES (?, ?, ?)", user.Id, user.Name, user.Bot)
		if err != nil {
			return err
		}
	}

	userList := UserList{}
	for _, user := range result {
		userList = append(userList, User{traq_uuid: user.Id, trap_id: user.Name, is_bot: user.Bot})
	}
	for i := 0; i < len(userList); i += 50 {
		_, err := db.NamedExec("INSERT INTO users (traq_uuid, trap_id, is_bot) VALUES (:traq_uuid, :trap_id, :is_bot)", userList[i:min(i+50, len(userList))])
		if err != nil {
			return err
		}
	}
	return nil
}

type User struct {
	traq_uuid string `db:"traq_uuid"`
	trap_id   string `db:"trap_id"`
	is_bot    bool   `db:"is_bot"`
}

type UserList []User

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
