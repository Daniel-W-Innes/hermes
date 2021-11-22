package controllers

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
)

func getDBMock() (*sql.DB, sqlmock.Sqlmock, *gorm.DB, error) {

	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, nil, err
	}

	mockedGorm, err := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{PrepareStmt: true})
	if err != nil {
		return nil, nil, nil, err
	}

	return mockDB, mock, mockedGorm, nil
}

func oracle(mockDB *sql.DB) (*gorm.DB, error) {
	return gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{PrepareStmt: true, DryRun: true})
}

func setup(t *testing.T) (sqlmock.Sqlmock, *gorm.DB, *gorm.DB, *models.Config) {
	mockDB, mock, db, err := getDBMock()
	if err != nil {
		t.Logf("failed to setup mock db %s\n", err)
		t.FailNow()
	}

	oracle, err := oracle(mockDB)
	if err != nil {
		t.Logf("failed to setup oracle %s\n", err)
		t.FailNow()
	}

	config, err := models.GetConfig()
	if err != nil {
		t.Logf("failed to get config %s\n", err)
		t.FailNow()
	}
	return mock, db, oracle, config
}

func TestAddUser(t *testing.T) {
	mock, db, oracle, config := setup(t)

	input := models.UserLogin{
		Username: "test_username",
		Password: "password1234",
	}

	stmt := oracle.Where("username = ?", input.Username).Limit(1).Find(&models.User{}).Statement

	mock.ExpectPrepare(stmt.SQL.String()).ExpectQuery().WithArgs(input.Username).WillReturnRows(sqlmock.NewRows([]string{"username"}))

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
	mock, db, oracle, config := setup(t)

	input := models.UserLogin{
		Username: "test_username",
		Password: "password1234",
	}

	stmt := oracle.Where("username = ?", input.Username).Limit(1).Find(&models.User{}).Statement

	mock.ExpectPrepare(stmt.SQL.String()).ExpectQuery().WithArgs(input.Username).WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow(input.Username))

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
	mock, db, oracle, config := setup(t)

	input := models.UserLogin{
		Username: "test_username",
		Password: "password1234",
	}

	stmt := oracle.Where("username = ?", input.Username).Limit(1).Find(&models.User{}).Statement

	mock.ExpectPrepare(stmt.SQL.String()).ExpectQuery().WithArgs(input.Username).WillReturnRows(sqlmock.NewRows([]string{"username"})).WillReturnError(gorm.ErrUnsupportedDriver)

	output, hermesErr := AddUser(db, config, &input)
	if hermesErr == nil || output != nil {
		t.Errorf("add user was expecting an err %s\n", output)
	} else if hermesErr.Error() != hermesErrors.InternalServerError("").Error() {
		t.Errorf("error is the wrong type %s\n", hermesErr.Error())
	} else if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("db expectations were not met %s\n", err)
	}
}

//
//func TestLogin(t *testing.T) {
//	mockDB, mock, db, err := getDBMock()
//	if err != nil {
//		t.FailNow()
//	}
//
//	oracle, err := oracle(mockDB)
//	if err != nil {
//		t.FailNow()
//	}
//
//	input := models.UserLogin{
//		Username: "test_username",
//		Password: "password1234",
//	}
//
//	stmt := oracle.Where("username = ?", input.Username).Select("password_key", "id").First(&models.User{}).Statement
//
//	mock.ExpectQuery(stmt.SQL.String()).WithArgs("test_username").WillReturnRows(sqlmock.NewRows([]string{"password_key", "id"}).AddRow("password1234", 1))
//
//	output, err := Login(db, &models.UserLogin{
//		Username: "test_username",
//		Password: "password1234",
//	})
//
//	log.Print(output)
//}
