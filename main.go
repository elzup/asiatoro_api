package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"

	"github.com/k0kubun/pp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	// User : Who are you
	User struct {
		ID    int64  `json:"id" form:"id" query:"id"`
		Name  string `json:"name" form:"name" query:"name"`
		Pass  string `json:"pass" form:"pass" query:"pass"`
		Token string `json:"token" form:"token" query:"token"`
	}

	// AccessPoint : Wifi AccessPoint profiles
	AccessPoint struct {
		ID    int64  `json:"id" form:"id" query:"id"`
		Ssid  string `json:"ssid" form:"ssid" query:"ssid"`
		Bssid string `json:"bssid" form:"bssid" query:"bssid"`
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
	followsTable      = "follows"
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

func selectFollow(u User, ap AccessPoint) Follow {
	var follow Follow
	sess.Select("*").
		From(followsTable).
		Where("user_id = ? AND access_point_id = ?", u.ID, ap.ID).
		Load(&follow)
	return follow
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
	result, err := sess.
		InsertInto(usersTable).
		Columns("name", "pass", "token").
		Values(u.Name, u.Pass, token).Exec()

	if err != nil {
		pp.Println(err)
		res := map[string]string{"message": err.Error()}
		return c.JSON(http.StatusConflict, res)
	}
	res, err := result.LastInsertId()
	u.ID = res
	u.Token = token
	return c.JSON(http.StatusCreated, u)
}

func createFollow(c echo.Context) error {
	u := c.Get("authorizedUser").(User)
	ap, err := findOrCreateAccessPoint(c)
	if err != nil {
		pp.Println(err)
		return err
	}
	follow := selectFollow(u, ap)
	if follow.User.ID == 0 {
		follow.User = u
		follow.AccessPoint = ap
		sess.
			InsertInto(followsTable).
			Columns("user_id", "access_point_id").
			Values(u.ID, ap.ID).Exec()
	}
	return c.JSON(http.StatusCreated, follow)
}

func findOrCreateAccessPoint(c echo.Context) (AccessPoint, error) {
	ap := new(AccessPoint)
	if err := c.Bind(ap); err != nil {
		return *ap, err
	}
	// TODO: primary only bssd
	sess.
		Select("*").
		From(accessPointsTable).
		Where("ssid = ? AND bssid = ?", ap.Ssid, ap.Bssid).
		LoadStruct(&ap)
	pp.Println(ap)
	if ap.ID == 0 {
		result, _ := sess.
			InsertInto(accessPointsTable).
			Columns("ssid", "bssid").
			Values(ap.Ssid, ap.Bssid).Exec()
		res, _ := result.LastInsertId()
		ap.ID = res
	}
	return *ap, nil
}

func oAuth2() echo.MiddlewareFunc {
	return middleware.KeyAuth(func(key string, c echo.Context) (error, bool) {
		var params = strings.SplitN(key, ":", 2)
		var id = params[0]
		var token = params[1]
		var u User
		sess.Select("*").From(usersTable).Where("id = ?", id).Load(&u)
		c.Set("authorizedUser", u)
		return nil, token == u.Token
	})
}

func main() {
	e := echo.New()

	e.POST("/users", createUser)
	e.POST("/follows", createFollow, oAuth2())

	e.Logger.Fatal(e.Start(":1323"))
}
