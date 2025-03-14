package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User model
type User struct {
	Id        string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string          `gorm:"not null" json:"name"`
	Email     string          `gorm:"not null" json:"email"`
	CreatedAt time.Time       `gorm:"default:now()" json:"created_at"`
	UpdatedAt *time.Time      `gorm:"default:null" json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index;default:null" json:"deleted_at"`
}

var db *gorm.DB

func main() {
	// Chama a função que carrega as variáveis de ambiente
	loadEnv()

	// Chama a função que conecta ao banco de dados
	setupDb()

	// Chama a função que configura o router do Gin
	router := setupRouter()

	// Inicia o servidor
	appPort := os.Getenv("API_PORT")
	fmt.Printf("Server running on port:%s", appPort)
	router.Run(fmt.Sprintf(":%s", appPort))
}

func loadEnv() {
	// Carrega as variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error on load .env file.")
	}
}

func setupDb() {
	// Acessa as variáveis de ambiente para conectar ao banco
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")

	// Conecta ao banco de dados
	dsn := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Fail on connect to database.")
	}

	// Cria a tabela automaticamente (se não existir)
	err = db.AutoMigrate(&User{})
	if err != nil {
		panic("Fail do migrate model User.")
	}
}

func setupRouter() *gin.Engine {
	// Configura o router do Gin
	router := gin.Default()

	// Rotas
	router.GET("/users", getUsers)
	router.GET("/users/:id", getUserById)
	router.POST("/users", postUser)
	router.PATCH("/users/:id", patchUser)
	router.DELETE("/users/:id", deleteUser)

	return router
}

// Retorna todos os usuários
func getUsers(c *gin.Context) {
	var users []User
	result := db.Where("deleted_at IS NULL").Find(&users)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, users)
}

// Retorna um usuário pelo Id
func getUserById(c *gin.Context) {
	id := c.Params.ByName("id")
	var user User
	result := db.Where("id = ? AND deleted_at IS NULL", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		} else {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		}
		return
	}
	c.IndentedJSON(http.StatusOK, user)
}

// Cria um novo usuário
func postUser(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	// Verifica se o usuário já existe
	var existingUser User
	result := db.Where("email = ? AND deleted_at IS NULL", newUser.Email).First(&existingUser)
	if result.RowsAffected > 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user already exists"})
		return
	}

	// Cria o usuário
	result = db.Create(&newUser)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newUser)
}

// Atualiza um usuário existente
func patchUser(c *gin.Context) {
	id := c.Params.ByName("id")
	var updates User
	if err := c.BindJSON(&updates); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}

	// Verifica se o usuário existe
	var existingUser User
	result := db.Where("id = ? AND deleted_at IS NULL", id).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}

	// Atualiza o usuário
	updates.UpdatedAt = &[]time.Time{time.Now()}[0]
	result = db.Model(&existingUser).Updates(updates)
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, existingUser)
}

// Deleta um usuário (soft delete)
func deleteUser(c *gin.Context) {
	id := c.Params.ByName("id")

	// Verifica se o usuário existe
	var existingUser User
	result := db.Where("id = ? AND deleted_at IS NULL", id).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}

	// Soft delete
	result = db.Model(&existingUser).Update("deleted_at", time.Now())
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "user deleted"})
}
