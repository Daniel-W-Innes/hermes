package controllers

import (
	"database/sql"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
)

const (
	getUserStatement = "SELECT * FROM \"users\" WHERE username = $1 AND \"users\".\"deleted_at\" IS NULL LIMIT 1"
	addUserStatement = "INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"username\",\"password_key\") VALUES ($1,$2,$3,$4,$5) RETURNING \"id\""
)

func getDBMock() (*sql.DB, sqlmock.Sqlmock, *gorm.DB, error) {

	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, nil, err
	}

	mockedGorm, err := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{PrepareStmt: true, SkipDefaultTransaction: true})
	if err != nil {
		return nil, nil, nil, err
	}

	return mockDB, mock, mockedGorm, nil
}

func setup(t *testing.T) (sqlmock.Sqlmock, *gorm.DB, *models.Config) {
	_, mock, db, err := getDBMock()
	if err != nil {
		t.Logf("failed to setup mock db %s\n", err)
		t.FailNow()
	}

	config, err := models.GetConfig()
	if err != nil {
		t.Logf("failed to get config %s\n", err)
		t.FailNow()
	}
	return mock, db, config
}

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type PasswordKey struct {
	Password []byte
	config   models.PasswordConfig
}

func (p PasswordKey) Match(v driver.Value) bool {
	user := models.User{
		PasswordKey: v.([]byte),
	}
	err := user.CheckPassword(&p.config, p.Password)
	return err == nil
}

func TestAddUser(t *testing.T) {
	mock, db, config := setup(t)

	input := models.UserLogin{
		Username: "test_username",
		Password: "password1234",
	}

	mock.ExpectPrepare(getUserStatement).ExpectQuery().WithArgs(input.Username).WillReturnRows(sqlmock.NewRows([]string{"username"}))

	mock.ExpectPrepare(addUserStatement).ExpectQuery().WithArgs(
		AnyTime{}, AnyTime{}, nil, input.Username, PasswordKey{Password: []byte(input.Password), config: config.PasswordConfig},
	).WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

	output, err := AddUser(db, config, &input)
	if err != nil {
		t.Errorf("add user return a unexpected err %s\n", err)
	} else if output.AccessToken == "" {
		t.Errorf("add user return a empty access token %s\n", err)
	} else if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("db expectations were not met %s\n", err)
	} else {
		re := regexp.MustCompile(`^[\w-]*\.[\w-]*\.[\w-]*$`)
		if !re.MatchString(output.AccessToken) {
			t.Logf("access token is not a jwt: %s", output.AccessToken)
			t.FailNow()
		}
	}
}

func TestAddUserUserExists(t *testing.T) {
	mock, db, config := setup(t)

	input := models.UserLogin{
		Username: "test_username",
		Password: "password1234",
	}

	mock.ExpectPrepare(getUserStatement).ExpectQuery().WithArgs(input.Username).WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow(input.Username))

	output, hermesErr := AddUser(db, config, &input)
	if hermesErr == nil || output != nil {
		t.Errorf("add user was expecting an err %s\n", output)
	} else if hermesErr.Error() != hermesErrors.UserExists().Error() {
		t.Errorf("error is the wrong type %s\n", hermesErr.Error())
	} else if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("db expectations were not met %s\n", err)
	}
}

func TestAddUserDBError(t *testing.T) {
	mock, db, config := setup(t)

	input := models.UserLogin{
		Username: "test_username",
		Password: "password1234",
	}

	mock.ExpectPrepare(getUserStatement).ExpectQuery().WithArgs(input.Username).WillReturnRows(sqlmock.NewRows([]string{"username"})).WillReturnError(gorm.ErrUnsupportedDriver)

	output, hermesErr := AddUser(db, config, &input)
	if hermesErr == nil || output != nil {
		t.Errorf("add user was expecting an err %s\n", output)
	} else if hermesErr.Error() != hermesErrors.InternalServerError("").Error() {
		t.Errorf("error is the wrong type %s\n", hermesErr.Error())
	} else if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("db expectations were not met %s\n", err)
	}
}
