package main

import (
	"log"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Leaderboard struct {
	gorm.Model
	Name  string `json:"name" validate:"required,min=3,max=100"`
	Score int    `json:"score" validate:"gte=0,lte=1000"`
}

var (
	db        *gorm.DB
	validate  *validator.Validate
	apiSecret string
)

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("leaderboard.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&Leaderboard{})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiSecret = os.Getenv("API_KEY")
	if apiSecret == "" {
		log.Fatal("API_KEY not set")
	}
	initDB()
	initAuth()
	validate = validator.New()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
	}))

	app.Use(func(c *fiber.Ctx) error {
		sess, err := Store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to get session",
				"details": err.Error(),
			})
		}
		c.Locals("session", sess)
		return c.Next()
	})

	app.Get("/login", login)
	app.Get("/callback", callback)
	app.Get("/user-info", userInfoHandler)
	app.Static("/", "./static")
	app.Get("/leaderboard", AuthMiddleware, getLeaderboard)

	api := app.Group("/leaderboard", APIKeyMiddleware)

	api.Post("/", createEntry)
	api.Put("/:id", updateEntry)
	api.Delete("/:id", deleteEntry)

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}

func APIKeyMiddleware(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "API key required"})
	}
	if apiKey != apiSecret {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid API key"})
	}
	return c.Next()
}

func AuthMiddleware(c *fiber.Ctx) error {
	sess, err := Store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}

	userInfo := sess.Get("user-info")
	if userInfo == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.Next()
}

func getLeaderboard(c *fiber.Ctx) error {
	var leaderboard []Leaderboard

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	result := db.Offset(offset).Limit(limit).Find(&leaderboard)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database query failed"})
	}

	var count int64
	db.Model(&Leaderboard{}).Count(&count)

	return c.JSON(fiber.Map{
		"data":       leaderboard,
		"page":       page,
		"limit":      limit,
		"total":      count,
		"totalPages": (count + int64(limit) - 1) / int64(limit),
	})
}

func createEntry(c *fiber.Ctx) error {
	var entry Leaderboard
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if err := validate.Struct(entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	result := db.Create(&entry)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create entry"})
	}
	return c.JSON(entry)
}

func updateEntry(c *fiber.Ctx) error {
	id := c.Params("id")
	var entry Leaderboard
	if err := db.First(&entry, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Entry not found"})
	}
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	if err := validate.Struct(entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	result := db.Save(&entry)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update entry"})
	}
	return c.JSON(entry)
}

func deleteEntry(c *fiber.Ctx) error {
	id := c.Params("id")
	var entry Leaderboard
	if err := db.First(&entry, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Entry not found"})
	}
	result := db.Delete(&entry)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete entry"})
	}
	return c.JSON(fiber.Map{"message": "Entry deleted"})
}
