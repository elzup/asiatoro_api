package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
	"github.com/k0kubun/pp"
	"github.com/labstack/echo"
)

type (
	// User : Who are you
	User struct {
		ID    int    `json:"id" form:"id" query:"id"`
		Name  string `json:"name" form:"name" query:"name"`
		Pass  string `json:"pass" form:"pass" query:"pass"`
		Token string `json:"token" form:"token" query:"token"`
	}

	// AccessPoint : Wifi AccessPoint profiles
	AccessPoint struct {
		Ssid  string
		Bssid string
	}

	// Follow : User and AccessPoint follow relation
	Follow struct {
		User        User
		AccessPoint AccessPoint
	}

	// Checkin : User checkins
	Checkin struct {
		User        User
		AccessPoint AccessPoint
		ts          dbr.NullTime
	}
)

var (
	usersTable        = "users"
	checkinsTable     = "logs"
	accessPointsTable = "access_points"
	seq               = 1
	conn, _           = dbr.Open("mysql", os.Getenv("MYSQL_URL"), nil)
	sess              = conn.NewSession(nil)
)

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func existsUser(u *User) bool {
	var r int64
	sess.Select("count(*)").From(usersTable).Where("name = ?", u.Name).Load(&r)
	return r > 0
}

func createUser(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	if u.Name == "" || u.Pass == "" {
		res := map[string]string{"message": "Name or Pass Field empty."}
		return c.JSON(http.StatusBadRequest, res)
	}
	if existsUser(u) {
		res := map[string]string{"message": "Duplicate user name."}
		return c.JSON(http.StatusBadRequest, res)
	}

	token := randToken()
	_, err := sess.
		InsertInto(usersTable).
		Columns("name", "pass", "token").
		Values(u.Name, u.Pass, token).Exec()
	if err != nil {
		pp.Print(err)
		res := map[string]string{"message": err.Error()}
		return c.JSON(http.StatusConflict, res)
	}
	u.Token = token
	return c.JSON(http.StatusCreated, u)
}

func selectLogs(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	token := randToken()
	sess.InsertInto(usersTable).Columns("name", "pass", "token").Values(u.Name, u.Pass, token).Exec()
	return c.JSON(http.StatusCreated, token)
}

func main() {
	e := echo.New()

	e.POST("/users", createUser)

	e.Logger.Fatal(e.Start(":1323"))
}
