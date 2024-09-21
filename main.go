package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique;not null"`
	Provider string `json:"provider"`
}
type QA struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	User      User   `json:"user" gorm:"foreignKey:UserEmail;references:Email;constraint:OnDelete:CASCADE"`
	UserEmail string `json:"user_email" gorm:"not null"`
}

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Failed to load: %v", err)
	}
}

func initDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"))
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&User{}, &QA{})
}

func main() {
	loadEnv()
	initDB()

	e := echo.New()

	e.GET("/api/items", getItemsByEmail)
	e.POST("/api/users", createUser)
	e.POST("/api/items", createItem)

	e.Logger.Fatal(e.Start(":8080"))
}

func getItemsByEmail(c echo.Context) error {
	var request struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if request.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	var items []QA
	if err := db.Preload("User").Where("user_email = ?", request.Email).Find(&items).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch items for the specified user"})
	}
	return c.JSON(http.StatusOK, items)
}

func createItem(c echo.Context) error {
	item := new(QA)
	if err := c.Bind(item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	if err := db.Create(&item).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create item"})
	}
	return c.JSON(http.StatusOK, item)
}

func createUser(c echo.Context) error {
	user := new(User)

	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}
	if result := db.Create(&user); result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create user"})
	}

	return c.JSON(http.StatusCreated, user)
}
