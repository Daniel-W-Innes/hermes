package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Daniel-W-Innes/messenger/models"
	"github.com/Daniel-W-Innes/messenger/utils"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"
	"time"
)

var userLogin = models.UserLogin{
	Username: "user",
	Password: "password",
}

func cleanDB() error {
	db, err := utils.Connection()
	if err != nil {
		return err
	}
	db.MustExec("TRUNCATE TABLE recipient RESTART IDENTITY CASCADE")
	db.MustExec("TRUNCATE TABLE message RESTART IDENTITY CASCADE")
	db.MustExec("TRUNCATE TABLE app_user RESTART IDENTITY CASCADE")
	return nil
}

func setup(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Logf("failed to init db %s", err)
		t.FailNow()
	}
}

func teardown(t *testing.T) {
	err := cleanDB()
	if err != nil {
		t.Logf("failed to clean db %s", err)
		t.FailNow()
	}
}

func getJwtFromResp(t *testing.T, resp *http.Response) string {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("failed to read body %s", err)
		t.FailNow()
	}
	var tokens models.JWT
	err = json.Unmarshal(b, &tokens)
	if err != nil {
		t.Logf("failed unmarshal body %s %s", string(b), err)
		t.FailNow()
	}
	return tokens.AccessToken
}

func assertJwtBody(t *testing.T, resp *http.Response) {
	if resp.StatusCode != fiber.StatusOK {
		t.Logf("bad status: %s", resp.Status)
		t.FailNow()
	} else if cType := resp.Header.Get(fiber.HeaderContentType); cType != fiber.MIMEApplicationJSON {
		t.Logf("bad Content-Type: %s", cType)
		t.FailNow()
	} else {
		accessToken := getJwtFromResp(t, resp)
		if accessToken == "" {
			t.Log("miss access_token")
			t.FailNow()
		} else {
			re := regexp.MustCompile(`^[\w-]*\.[\w-]*\.[\w-]*$`)
			if !re.MatchString(accessToken) {
				t.Logf("access token is not a jwt: %s", accessToken)
				t.FailNow()
			}
		}
	}
}

func addUser(t *testing.T, app *fiber.App) *http.Response {
	reqBodyBytes, err := json.Marshal(userLogin)
	if err != nil {
		t.Log(fmt.Errorf("failed to marshal body %w", err))
		t.FailNow()
	}
	req := httptest.NewRequest("POST", "/user", bytes.NewReader(reqBodyBytes))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req, int(time.Hour.Milliseconds()))
	if err != nil {
		t.Logf("failed to test app %s", err)
		t.FailNow()
	}
	return resp
}

func TestAddUser(t *testing.T) {
	app := getApp()

	setup(t)
	defer teardown(t)

	resp := addUser(t, app)

	db, err := utils.Connection()
	if err != nil {
		t.Log("failed to connect to db")
		t.FailNow()
	}

	var user models.User
	err = db.Get(&user, "SELECT * FROM app_user WHERE username=$1", userLogin.Username)
	if err != nil {
		t.Logf("user missing from db: %s", err)
		t.FailNow()
	} else {
		re := regexp.MustCompile(`(?m)^\$2[ayb]\$.{56}$`)
		if !re.MatchString(string(user.PasswordKey)) {
			t.Logf("password key is not hashed: %s", string(user.PasswordKey))
			t.FailNow()
		}
		assertJwtBody(t, resp)
	}
}

func TestLogin(t *testing.T) {
	app := getApp()

	setup(t)
	defer teardown(t)

	addUser(t, app)

	reqBodyBytes, err := json.Marshal(userLogin)
	if err != nil {
		t.Log(fmt.Errorf("failed to marshal body %w", err))
		t.FailNow()
	}
	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(reqBodyBytes))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req, int(time.Hour.Milliseconds()))
	if err != nil {
		t.Logf("failed to test app %s", err)
		t.FailNow()
	}
	assertJwtBody(t, resp)
}

func addMessage(t *testing.T, app *fiber.App, token string, message map[string]map[string]interface{}) *http.Response {

	reqBodyBytes, err := json.Marshal(message)
	if err != nil {
		t.Logf("failed to marshal body %s", err)
		t.FailNow()
	}

	req := httptest.NewRequest("POST", "/message", bytes.NewReader(reqBodyBytes))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

	resp, err := app.Test(req, int(time.Hour.Milliseconds()))
	if err != nil {
		t.Logf("failed to test app %s", err)
		t.FailNow()
	}

	return resp
}

func TestAddMessage(t *testing.T) {
	app := getApp()

	setup(t)
	defer teardown(t)
	resp := addUser(t, app)
	token := getJwtFromResp(t, resp)

	resp = addMessage(t, app, token, map[string]map[string]interface{}{"message": {"text": "test"}})

	db, err := utils.Connection()
	if err != nil {
		t.Log("failed to connect to db")
		t.FailNow()
	}

	var message models.Message
	err = db.Get(&message, "SELECT * FROM message WHERE id=1")
	if err != nil {
		t.Logf("message missing from db: %s", err)
		t.FailNow()
	}

	if message.Text != "test" || message.ID != 1 {
		t.Logf("the message is not the same %s", message.Text)
		t.FailNow()
	}

	if message.Palindrome {
		t.Logf("the message not palindrome %s", message.Text)
		t.FailNow()
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Logf("bad status: %s", resp.Status)
		t.FailNow()
	} else if cType := resp.Header.Get(fiber.HeaderContentType); cType != fiber.MIMEApplicationJSON {
		t.Logf("bad Content-Type: %s", cType)
		t.FailNow()
	} else {
		output := make(map[string]int)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Logf("failed to read body %s", err)
			t.FailNow()
		}
		err = json.Unmarshal(b, &output)
		if err != nil {
			t.Logf("failed unmarshal body %s %s", string(b), err)
			t.FailNow()
		}
		if output["id"] != 1 {
			t.Logf("bad id in response %v", output)
			t.FailNow()
		}
	}
}

func TestGetMessage(t *testing.T) {
	app := getApp()

	setup(t)
	defer teardown(t)

	resp := addUser(t, app)
	token := getJwtFromResp(t, resp)

	_ = addMessage(t, app, token, map[string]map[string]interface{}{"message": {"text": "test"}})

	req := httptest.NewRequest("GET", "/message?id=1", nil)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

	resp, err := app.Test(req, int(time.Hour.Milliseconds()))
	if err != nil {
		t.Logf("failed to test app %s", err)
		t.FailNow()
	} else if resp.StatusCode != fiber.StatusOK {
		t.Logf("bad status: %s", resp.Status)
		t.FailNow()
	} else if cType := resp.Header.Get(fiber.HeaderContentType); cType != fiber.MIMEApplicationJSON {
		t.Logf("bad Content-Type: %s", cType)
		t.FailNow()
	} else {
		var messageBundle models.MessageBundle
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Logf("failed to read body %s", err)
			t.FailNow()
		} else if err = json.Unmarshal(b, &messageBundle); err != nil {
			t.Logf("failed unmarshal body %s %s", string(b), err)
			t.FailNow()
		} else {
			if !reflect.DeepEqual(models.MessageBundle{
				Message: models.Message{
					ID:         1,
					OwnerID:    1,
					Text:       "test",
					Palindrome: false,
				},
				RecipientIds: nil,
			}, messageBundle) {
				t.Logf("the message is not the same %s", messageBundle.Message.Text)
				t.FailNow()
			}
		}
	}
}

func TestGetMessages(t *testing.T) {
	app := getApp()

	setup(t)
	defer teardown(t)

	resp := addUser(t, app)
	token := getJwtFromResp(t, resp)

	numberMessage := 2

	for i := 1; i <= numberMessage; i++ {
		_ = addMessage(t, app, token, map[string]map[string]interface{}{"message": {"text": fmt.Sprintf("test%d", i)}})
	}

	req := httptest.NewRequest("GET", "/message", nil)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

	resp, err := app.Test(req, int(time.Hour.Milliseconds()))
	if err != nil {
		t.Logf("failed to test app %s", err)
		t.FailNow()
	} else if resp.StatusCode != fiber.StatusOK {
		t.Logf("bad status: %s", resp.Status)
		t.FailNow()
	} else if cType := resp.Header.Get(fiber.HeaderContentType); cType != fiber.MIMEApplicationJSON {
		t.Logf("bad Content-Type: %s", cType)
		t.FailNow()
	} else {
		messages := new(map[string][]models.MessageBundle)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Logf("failed to read body %s", err)
			t.FailNow()
		} else if err = json.Unmarshal(b, &messages); err != nil {
			t.Logf("failed unmarshal body %s %s", string(b), err)
			t.FailNow()
		} else {
			for i := 1; i <= 2; i++ {
				if !reflect.DeepEqual(models.MessageBundle{
					Message: models.Message{
						ID:         i,
						OwnerID:    1,
						Text:       fmt.Sprintf("test%d", i),
						Palindrome: false,
					},
					RecipientIds: nil,
				}, (*messages)["messages"][i-1]) {
					t.Logf("the message is not the same %v", (*messages)["messages"][i-1])
					t.FailNow()
				}
			}
		}
	}
}
func TestDeleteMessage(t *testing.T) {
	app := getApp()

	setup(t)
	defer teardown(t)

	resp := addUser(t, app)
	token := getJwtFromResp(t, resp)

	_ = addMessage(t, app, token, map[string]map[string]interface{}{"message": {"text": "test"}})

	req := httptest.NewRequest("DELETE", "/message?id=1", nil)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

	resp, err := app.Test(req, int(time.Hour.Milliseconds()))
	if err != nil {
		t.Logf("failed to test app %s", err)
		t.FailNow()
	} else if resp.StatusCode != fiber.StatusOK {
		t.Logf("bad status: %s", resp.Status)
		t.FailNow()
	} else if cType := resp.Header.Get(fiber.HeaderContentType); cType != fiber.MIMETextPlainCharsetUTF8 {
		t.Logf("bad Content-Type: %s", cType)
		t.FailNow()
	} else {
		db, err := utils.Connection()
		if err != nil {
			t.Log("failed to connect to db")
			t.FailNow()
		}

		var message models.Message
		err = db.Get(&message, "SELECT * FROM message WHERE id=1")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Logf("message not removed db: %v", message)
			t.FailNow()
		}

	}
}
