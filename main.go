package main

// golang gets also its packages from github.
// like js does it from npm

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq" // postgres driver
)

// create models using struct
type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// 1234

func main() {
	e := echo.New()
	// loging package, middleware
	e.Use(middleware.RequestLogger())
	// database connect karna h
	// database connectivity string // connection string

	// just trying out.
	dsn := "host = localhost port=5432 user=postgres password=1234 dbname=newdb password=1234"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to connect database -----  ", err)
	}
	// close db., defer runs when the entire function is about to terminate, good practice to close all existing connections created.
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("database ping failed ---- ", err)
	}
	// how to create table.
	createTable := `
	   create table if not exists users (
	     id serial primary key,
		 name text,
		 email text unique,
		 age int   
	   );
	`
	// creates output on the terminal, used for displaying errors on the terminals.
	if _, err := db.Exec(createTable); err != nil {
		log.Fatal("Failed to created the table ---- ", err)
	}
	e.POST("/users", func(c echo.Context) error {
		u := new(User)
		if err := c.Bind(&u); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Request Body"})
		}
		var id int
		//$1, $2, $3 it is similar to how we used to do in bash scripting passing arguments to functions in bash.
		err := db.QueryRow(
			"Insert into users (name, email, age) values ($1, $2, $3) returning id", u.Name, u.Email, u.Age,
		).Scan(&id)

		if err != nil {
			// takes a httpcode to return and an interface.
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		u.Id = id
		return c.JSON(http.StatusCreated, u)
	})

	e.GET("/users", func(c echo.Context) error {
		rows, err := db.Query("SELECT id,name,email,age from users")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()
		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.Id, &u.Name, &u.Email, &u.Age); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			users = append(users, u)
		}
		return c.JSON(http.StatusOK, users)
	})

	e.GET("/users/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}
		var u User
		err = db.QueryRow("SELECT id,name,email,age from users where id = $1", id).Scan(&u.Id, &u.Name, &u.Email, &u.Age)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, u)
	})

	e.PUT("/users/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		// when using put - the entire row gets replaced with the data you send, truly what it does, is creates a row of your data, remove the existing data present on that id, and insert this new data in place of that
		// while using patch - the columns you pass in the body, they only get updated.
		u := new(User)
		if err := c.Bind(&u); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Request Body"})
		}
		result, err := db.Exec("Update users set name = $1, email = $2, age = $3 where id = $4", u.Name, u.Email, u.Age, id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		u.Id = id
		return c.JSON(http.StatusOK, u)
	})
	e.Logger.Fatal(e.Start(":8090"))
}
