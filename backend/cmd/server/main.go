package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/ihsanbudiman/throtl/internal/handler"
	"github.com/ihsanbudiman/throtl/internal/middleware"
	"github.com/ihsanbudiman/throtl/internal/proxy"
	"github.com/ihsanbudiman/throtl/internal/store"
)

func main() {
	dbURL := os.Getenv("THROTL_DB_URL")
	if dbURL == "" {
		dbURL = "throtl.db"
	}
	port := os.Getenv("THROTL_PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("THROTL_JWT_SECRET")
	if jwtSecret == "" {
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			log.Fatalf("Failed to generate JWT secret: %v", err)
		}
		jwtSecret = hex.EncodeToString(bytes)
		fmt.Println("⚠️  No THROTL_JWT_SECRET set — using auto-generated secret (tokens invalidate on restart)")
	}

	// Init store
	s, err := store.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer s.Close()

	// Init components
	rl := middleware.NewRateLimiter(s)
	h := handler.New(s, jwtSecret, rl)
	p := proxy.NewOpenAIProxy(s)

	// Echo server
	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS, echo.PATCH},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders: []string{"Retry-After"},
	}))

	// --- Public auth routes (no JWT required) ---
	auth := e.Group("/api/auth")
	auth.GET("/check", h.CheckSetup)
	auth.POST("/setup", h.Setup)
	auth.POST("/login", h.Login)

	// --- Protected dashboard API (JWT required) ---
	api := e.Group("/api", middleware.JWTAuth([]byte(jwtSecret)))
	api.GET("/me", h.GetMe)
	api.GET("/stats", h.GetStats)
	api.GET("/usage", h.GetUsageLogs)

	// Providers
	api.GET("/providers", h.ListProviders)
	api.POST("/providers", h.CreateProvider)
	api.DELETE("/providers/:id", h.DeleteProvider)

	// API Keys
	api.GET("/keys", h.ListAPIKeys)
	api.POST("/keys", h.CreateAPIKey)
	api.PATCH("/keys/:id", h.ToggleAPIKey)
	api.DELETE("/keys/:id", h.DeleteAPIKey)

	// Models
	api.GET("/models", h.ListModels)
	api.PATCH("/models/:id", h.ToggleModel)

	// --- Proxy API (consumer-facing, share-key auth) ---
	proxyGroup := e.Group("/v1", middleware.KeyAuth(s), rl.Middleware())
	proxyGroup.GET("/models", p.ListModels)
	proxyGroup.Any("/*", p.ProxyHandler)

	e.Any("/v1", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "Throtl Gateway — Use /v1/chat/completions, /v1/models, etc.",
		})
	})

	fmt.Printf("🚀 Throtl Gateway running on http://localhost:%s\n", port)
	fmt.Printf("   Dashboard API:  http://localhost:%s/api/*\n", port)
	fmt.Printf("   Proxy endpoint: http://localhost:%s/v1/*\n", port)
	e.Logger.Fatal(e.Start(":" + port))
}
