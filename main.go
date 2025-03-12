package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type user struct {
	Id        string     `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	// Avatar  string `json:"avatar"`
}

func getUserByID(c *gin.Context) {
	db := OpenConn()
	defer db.Close()

	id := c.Params.ByName("id")

	var existsUser user

	sql := `SELECT * FROM public.user WHERE id = $1 AND deleted_at IS NULL`
	query := db.QueryRow(sql, id)
	query.Scan(
		&existsUser.Id,
		&existsUser.Name,
		&existsUser.Email,
		&existsUser.CreatedAt,
		&existsUser.UpdatedAt,
		&existsUser.DeletedAt,
	)

	if err := query.Err(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if existsUser.Id == "" {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, existsUser)
}

func getUsers(c *gin.Context) {
	db := OpenConn()
	defer db.Close()

	var users []user = []user{}

	sql := `SELECT * FROM public.user WHERE deleted_at IS NULL`
	query, err := db.Query(sql)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	defer query.Close()

	for query.Next() {
		var u user

		err = query.Scan(
			&u.Id,
			&u.Name,
			&u.Email,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		users = append(users, u)
	}

	c.IndentedJSON(http.StatusOK, users)
}

func postUser(c *gin.Context) {
	db := OpenConn()
	defer db.Close()

	var newUser user

	if err := c.BindJSON(&newUser); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	sqlExists := `SELECT id FROM public.user WHERE email = $1 AND deleted_at IS NULL`
	queryExists, err := db.Exec(sqlExists, newUser.Email)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if row, _ := queryExists.RowsAffected(); row > 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user already exists"})
		return
	}

	sql := `INSERT INTO public.user (name, email) VALUES ($1, $2) RETURNING id, created_at, updated_at, deleted_at`

	query := db.QueryRow(sql, newUser.Name, newUser.Email)
	query.Scan(
		&newUser.Id,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
		&newUser.DeletedAt,
	)

	if err := query.Err(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newUser)
}

func patchUser(c *gin.Context) {
	db := OpenConn()
	defer db.Close()

	id := c.Params.ByName("id")

	var existsUser user

	if err := c.BindJSON(&existsUser); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	sqlExists := `SELECT id FROM public.user WHERE id = $1 AND deleted_at IS NULL`
	queryExists, err := db.Exec(sqlExists, id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if row, _ := queryExists.RowsAffected(); row == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	sql := `UPDATE public.user SET name = $1, updated_at = now() WHERE id = $2 RETURNING id, email, created_at, updated_at, deleted_at`
	query := db.QueryRow(sql, existsUser.Name, id)
	query.Scan(
		&existsUser.Id,
		&existsUser.Email,
		&existsUser.CreatedAt,
		&existsUser.UpdatedAt,
		&existsUser.DeletedAt,
	)

	if err := query.Err(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, existsUser)
}

func deleteUser(c *gin.Context) {
	db := OpenConn()
	defer db.Close()

	id := c.Params.ByName("id")

	var existsUser user = user{Id: id}

	sqlExists := `SELECT id FROM public.user WHERE id = $1 AND deleted_at IS NULL`
	queryExists, err := db.Exec(sqlExists, id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if row, _ := queryExists.RowsAffected(); row == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	sql := `UPDATE public.user SET deleted_at = now() WHERE id = $1 RETURNING id, email, name, created_at, updated_at, deleted_at`
	query := db.QueryRow(sql, id)
	query.Scan(
		&existsUser.Id,
		&existsUser.Email,
		&existsUser.Name,
		&existsUser.CreatedAt,
		&existsUser.UpdatedAt,
		&existsUser.DeletedAt,
	)

	if err := query.Err(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, existsUser)
}

func OpenConn() *sql.DB {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=go-api sslmode=disable")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func main() {
	router := gin.Default()

	router.GET("/users", getUsers)
	router.GET("/users/:id", getUserByID)
	router.POST("/users", postUser)
	router.PATCH("/users/:id", patchUser)
	router.DELETE("/users/:id", deleteUser)

	fmt.Println("Server running on port 8080")
	router.Run("localhost:8080")
}
