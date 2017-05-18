package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
	"github.com/labstack/echo"
)

type (
	// User : Who are you
	User struct {
		Name  string
		Pass  string
		Token string
	}

	// UserPost : User request format
	UserPost struct {
		Name string `json:"name"`
		Pass string `json:"pass"`
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

func createUser(c echo.Context) error {
	u := new(UserPost)
	if err := c.Bind(u); err != nil {
		return err
	}
	token := randToken()
	fmt.Println(u)
	fmt.Println(u.Name)
	fmt.Println(u.Pass)
	result, err := sess.
		InsertInto(usersTable).
		Columns("name", "pass", "token").
		Values(u.Name, u.Pass, token).Exec()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
	response := map[string]string{"token": token}
	return c.JSON(http.StatusCreated, response)
}

func selectLogs(c echo.Context) error {
	u := new(UserPost)
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
