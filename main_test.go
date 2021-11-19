package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/Daniel-W-Innes/hermes/utils"
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
	db.Exec("TRUNCATE TABLE recipients RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE messages RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
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

	assertJwtBody(t, resp)

	user := new(models.User)
	result := db.Where("username = ?", userLogin.Username).First(user)
	if result.Error != nil {
		t.Logf("user missing from db: %s", err)
		t.FailNow()
	} else {
		re := regexp.MustCompile(`(?m)^\$2[ayb]\$.{56}$`)
		if !re.MatchString(string(user.PasswordKey)) {
			t.Logf("password key is not hashed: %s", string(user.PasswordKey))
			t.FailNow()
		}
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

func addMessage(t *testing.T, app *fiber.App, token string, message map[string]interface{}) *http.Response {

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

	resp = addMessage(t, app, token, map[string]interface{}{"text": "test"})

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

	db, err := utils.Connection()
	if err != nil {
		t.Log("failed to connect to db")
		t.FailNow()
	}
	var messages []models.Message
	result := db.Find(&messages)
	if result.Error != nil {
		t.Logf("failed to query db: %s", result.Error)
		t.FailNow()
	}
	if result.RowsAffected != 1 {
		t.Logf("wrong number of rows in db: %d", result.RowsAffected)
		t.FailNow()
	}

	if messages[0].Text != "test" || messages[0].ID != 1 {
		t.Logf("the message is not the same %s", messages[0].Text)
		t.FailNow()
	}

	if messages[0].Palindrome {
		t.Logf("the message not palindrome %s", messages[0].Text)
		t.FailNow()
	}
}

func TestGetMessage(t *testing.T) {
	app := getApp()

	setup(t)
	defer teardown(t)

	resp := addUser(t, app)
	token := getJwtFromResp(t, resp)

	_ = addMessage(t, app, token, map[string]interface{}{"text": "test"})

	req := httptest.NewRequest("GET", "/message/1", nil)
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
		var message models.Message
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Logf("failed to read body %s", err)
			t.FailNow()
		} else if err = json.Unmarshal(b, &message); err != nil {
			t.Logf("failed unmarshal body %s %s", string(b), err)
			t.FailNow()
		} else {
			if !reflect.DeepEqual(models.Message{
				ID:         1,
				OwnerID:    1,
				Text:       "test",
				Palindrome: false,
			}, message) {
				t.Logf("the message is not the same %s", message.Text)
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
		_ = addMessage(t, app, token, map[string]interface{}{"text": fmt.Sprintf("test%d", i)})
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
		var messages map[string][]models.Message
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Logf("failed to read body %s", err)
			t.FailNow()
		} else if err = json.Unmarshal(b, &messages); err != nil {
			t.Logf("failed unmarshal body %s %s", string(b), err)
			t.FailNow()
		} else if len(messages["messages"]) != numberMessage {
			t.Logf("wrong number of messages in body %s", string(b))
			t.FailNow()
		} else {
			for i := 1; i <= numberMessage; i++ {
				if !reflect.DeepEqual(models.Message{
					ID:         uint(i),
					OwnerID:    1,
					Text:       fmt.Sprintf("test%d", i),
					Palindrome: false,
				}, messages["messages"][i-1]) {
					t.Logf("the message is not the same %v", messages["messages"][i-1])
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

	_ = addMessage(t, app, token, map[string]interface{}{"text": "test"})

	req := httptest.NewRequest("DELETE", "/message/1", nil)
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
		db, err := utils.Connection()
		if err != nil {
			t.Log("failed to connect to db")
			t.FailNow()
		}

		var messages []models.Message
		result := db.Find(&messages)
		if result.Error != nil {
			t.Logf("failed to query db: %s", result.Error)
			t.FailNow()
		}
		if result.RowsAffected != 0 {
			t.Logf("wrong number of rows in db: %d", result.RowsAffected)
			t.FailNow()
		}
	}
}
