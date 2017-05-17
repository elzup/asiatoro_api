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

	UserPost struct {
		Name string `json:"name"`
		Pass string `json:"pass"`
	}
)

var (
	tablename = "User"
	seq       = 1
	conn, _   = dbr.Open("mysql", os.Getenv("MYSQL_URL"), nil)
	sess      = conn.NewSession(nil)
)

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/users", func(c echo.Context) error {
		u := new(UserPost)
		if err := c.Bind(u); err != nil {
			return err
		}
		token := randToken()
		sess.InsertInto(tablename).Columns("name", "pass", "token").Values(u.Name, u.Pass, token).Exec()
		return c.JSON(http.StatusCreated, token)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
