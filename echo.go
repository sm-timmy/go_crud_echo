package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"strconv"
	"time"
)

type user struct {
	Id   int    `param:"id" query:"id" form:"id" json:"id"`
	Name string `param:"name" query:"name" form:"name" json:"name"`
	Age  int    `param:"age" query:"age" form:"age" json:"age"`
}

type (
	dbContext struct {
		echo.Context
		db *sqlx.DB
	}
)

func main() {
	db, err := GetDB("postgres://postgres:postgres@localhost:5432/postgres") //or: db, err := GetDB(os.Getenv("PG_URI"))
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("DB connection success")

	var myUser = user{}
	db.Get(&myUser, "SELECT * from users")
	log.Println("My user :", myUser)

	e := echo.New()
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &dbContext{c, db}
			return h(cc)
		}
	})

	e.GET("/users", getUsers)
	e.POST("/users", saveUser)
	e.GET("/users/:id", getUser)
	e.PUT("/users/", updateUser)
	e.DELETE("/users/:id", deleteUser)
	e.Logger.Fatal(e.Start(":1323"))
}

// e.PUT("/users/", updateUser)
func updateUser(c echo.Context) error {
	cc := c.(*dbContext)
	u := new(user)
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	_, err := cc.db.Query("UPDATE users SET name = $1, age = $2 WHERE id = $3", u.Name, u.Age, u.Id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, u)
}

// e.POST("/users", saveUser)
func saveUser(c echo.Context) error {
	cc := c.(*dbContext)
	u := new(user)
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	_, err := cc.db.Query("INSERT into users VALUES (DEFAULT, $1, $2)", u.Name, u.Age)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, u)
}

// e.DELETE("/users/:id", deleteUser)
func deleteUser(c echo.Context) error {
	cc := c.(*dbContext)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	_, err = cc.db.Query("DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, "Ok")
}

// e.GET("/users/:id", getUser)
func getUser(c echo.Context) error {
	cc := c.(*dbContext)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	var name string
	var age int

	rows, err := cc.db.Query("SELECT * FROM users WHERE id=$1", id)
	if err != nil {
		return err
	}
	if rows.Next() {
		err := rows.Scan(&id, &name, &age)
		if err != nil {
			return err
		}
	} else {
		return c.JSON(http.StatusNotFound, "User not found")
	}
	user := user{id, name, age}
	return c.JSON(http.StatusOK, user)
}

// e.GET("/", getUsers)
func getUsers(c echo.Context) error {
	cc := c.(*dbContext)
	var users []user
	var id int
	var name string
	var age int

	rows, err := cc.db.Query("SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &name, &age)
		if err != nil {
			return err
		}
		users = append(users, user{id, name, age})
	}
	return c.JSON(http.StatusOK, users)
}

func GetDB(uri string) (*sqlx.DB, error) {
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
			 CREATE TABLE IF NOT EXISTS users(
			 	id SERIAL,
				name varchar NOT NULL,
				age int
			 );
		`)
		if err != nil {
			return err
		}
		return nil
	})
	pgxdb := stdlib.OpenDB(*connConfig, afterConnect)
	return sqlx.NewDb(pgxdb, "pgx"), nil
}
