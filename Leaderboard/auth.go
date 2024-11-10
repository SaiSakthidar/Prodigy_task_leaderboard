package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

var (
	OAuthConfig *oauth2.Config
	Store       *session.Store
)

func initAuth() {

	OAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	Store = session.New(session.Config{
		KeyLookup:      "cookie:sessionid",
		Expiration:     24 * time.Hour,
		CookieHTTPOnly: true,
		CookieSecure:   false,
		CookieSameSite: "lax",
	})

	gob.Register(UserInfo{})
}

func login(c *fiber.Ctx) error {
	url := OAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

func callback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Code not found in query"})
	}

	token, err := OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to exchange token",
			"details": err.Error(),
		})
	}

	client := OAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get user info",
			"details": err.Error(),
		})
	}
	defer resp.Body.Close()

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to decode user info",
			"details": err.Error(),
		})
	}

	sess, err := Store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get session",
			"details": err.Error(),
		})
	}

	sess.Set("user-info", userInfo)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":    "Failed to save session",
			"details":  err.Error(),
			"location": "callback function",
		})
	}

	return c.Redirect("/welcome.html")
}

func userInfoHandler(c *fiber.Ctx) error {
	sess, err := Store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}

	userInfo := sess.Get("user-info")
	if userInfo == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No user info found in session"})
	}
	user, ok := userInfo.(UserInfo)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid session data"})
	}

	return c.JSON(user)
}
