package user

import (
	"maryan_api/internal/infrastructure/db"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newTestUserRepoMysql(t *testing.T) (userRepoMySQL, sqlmock.Sqlmock, func() error) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatal(err.Error())
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})

	if err != nil {
		t.Fatal(err.Error())
	}

	return userRepoMySQL{gormDB}, mock, db.Close

}

func Test_userRepoMySQL_repo(t *testing.T) {
	repo, _, close := newTestUserRepoMysql(t)
	defer close()

	if &repo != repo.repo() {
		t.Error("Expected result to be a pointer to the repository.")
	}
}

func Test_userRepoMySQL_database(t *testing.T) {
	repo, _, close := newTestUserRepoMysql(t)
	defer close()

	if repo.db != repo.database() {
		t.Error("Expected result to be a pointer to repository db.")
	}
}

func Test_userRepoImpl_login(t *testing.T) {
	db, mock := db.NewMockDB()

	id := uuid.NewString()
	rows := sqlmock.NewRows([]string{
		"id",
		"email",
		"password",
	}).AddRow(
		id,
		"john.doe@example.com",
		"$2a$12$abcdef1234567890abcdef1234567890abcdef",
	)

	mock.ExpectQuery("SELECT `id`,`password` FROM `users` WHERE email = ?").WillReturnRows(rows)

	repo := userRepoMySQL{db}
	resultID, resultPassword, err := repo.login("john.doe@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if resultID.String() != id {
		t.Error(resultID.String())
	}

	if resultPassword != "$2a$12$abcdef1234567890abcdef1234567890abcdef" {
		t.Error(resultPassword)
	}

}
