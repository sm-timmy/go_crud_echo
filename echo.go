package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
)

type user struct {
	Id   int
	Name string
	Age  int
}

func main() {
	db, err := GetDB("postgres://postgres:postgres@localhost:5432/postgres")
	//db, err := GetDB(os.Getenv("PG_URI"))
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Worked!!")

	var myUser = user{}
	var id int
	var somekey string
	db.Get(&myUser, "SELECT * from users")
	db.Get(&somekey, "SELECT current_setting('some.key')")
	log.Println("My user :", myUser)
	log.Println("sometable id:", id)
	log.Println("somekey value:", somekey)

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	//e.POST("/users", saveUser)
	e.GET("/users/:id", getUser)
	//e.PUT("/users/:id", updateUser)
	//e.DELETE("/users/:id", deleteUser)
	e.Logger.Fatal(e.Start(":1323"))
}

func GetDB(uri string) (*sqlx.DB, error) {
	// before : directly using sqlx
	// DB, err = sqlx.Connect("postgres", uri)
	// after : using pgx to setup connection
	DB, err := PgxCreateDB(uri)
	if err != nil {
		return nil, err
	}
	DB.SetMaxIdleConns(2)
	DB.SetMaxOpenConns(4)
	DB.SetConnMaxLifetime(time.Duration(30) * time.Minute)

	return DB, nil
}

func PgxCreateDB(uri string) (*sqlx.DB, error) {
	connConfig, _ := pgx.ParseConfig(uri)
	afterConnect := stdlib.OptionAfterConnect(func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, `
			 SET SESSION "some.key" = 'somekey';
			 CREATE TEMP TABLE IF NOT EXISTS sometable AS SELECT 212 id;
		`)
		if err != nil {
			return err
		}
		return nil
	})

	pgxdb := stdlib.OpenDB(*connConfig, afterConnect)
	return sqlx.NewDb(pgxdb, "pgx"), nil
}

// e.GET("/users/:id", getUser)
func getUser(c echo.Context) error {
	// User ID from path `users/:id`
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

// e.POST("/save", save)
func save(c echo.Context) error {
	// Get name and email
	name := c.FormValue("name")
	email := c.FormValue("email")
	return c.String(http.StatusOK, "name:"+name+", email:"+email)
}
