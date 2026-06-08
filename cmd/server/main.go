package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/tejdeep/devfolio/internal/config"
	"github.com/tejdeep/devfolio/internal/db"
	"github.com/tejdeep/devfolio/internal/handlers"
	"github.com/tejdeep/devfolio/internal/middleware"
	"github.com/tejdeep/devfolio/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	// ── Infrastructure ─────────────────────────────────────────────────────
	pgPool, err := db.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("PostgreSQL: %v", err)
	}
	defer pgPool.Close()
	log.Println("✓ PostgreSQL connected")

	rdb := db.NewRedisClient(cfg.RedisURL)
	defer rdb.Close()
	log.Println("✓ Redis connected")

	// ── Repositories ──────────────────────────────────────────────────────
	projectRepo    := repository.NewProjectRepository(pgPool)
	skillRepo      := repository.NewSkillRepository(pgPool)
	experienceRepo := repository.NewExperienceRepository(pgPool)
	messageRepo    := repository.NewMessageRepository(pgPool)
	adminRepo      := repository.NewAdminRepository(pgPool)

	// ── Seed admin account ────────────────────────────────────────────────
	if err := seedAdmin(adminRepo, cfg); err != nil {
		log.Printf("Admin seed warning: %v", err)
	}

	// ── Middleware & Handlers ─────────────────────────────────────────────
	authMW         := middleware.NewAuthMiddleware(cfg)
	portfolioH     := handlers.NewPortfolioHandler(projectRepo, skillRepo, experienceRepo, rdb)
	contactH       := handlers.NewContactHandler(messageRepo)
	adminAuthH     := handlers.NewAdminAuthHandler(adminRepo, cfg)
	adminContentH  := handlers.NewAdminContentHandler(projectRepo, skillRepo, experienceRepo, messageRepo)

	// ── Router ─────────────────────────────────────────────────────────────
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Serve frontend
	r.Static("/static", "./web/static")
	r.GET("/", func(c *gin.Context) { c.File("./web/static/index.html") })
	r.GET("/admin", func(c *gin.Context) { c.File("./web/static/admin.html") })

	// ── Public API ─────────────────────────────────────────────────────────
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "time": time.Now()})
		})
		api.GET("/portfolio", portfolioH.GetAll)
		api.GET("/projects", portfolioH.GetProjects)
		api.GET("/skills", portfolioH.GetSkills)
		api.GET("/experience", portfolioH.GetExperience)
		api.POST("/contact", contactH.Submit)
		api.POST("/admin/login", adminAuthH.Login)
	}

	// ── Admin API (JWT protected) ──────────────────────────────────────────
	admin := r.Group("/api/admin")
	admin.Use(authMW.RequireAdmin())
	{
		// Projects
		admin.POST("/projects/:id", adminContentH.UpsertProject)
		admin.DELETE("/projects/:id", adminContentH.DeleteProject)

		// Skills
		admin.POST("/skills/:id", adminContentH.UpsertSkill)
		admin.DELETE("/skills/:id", adminContentH.DeleteSkill)

		// Experience
		admin.POST("/experience/:id", adminContentH.UpsertExperience)
		admin.DELETE("/experience/:id", adminContentH.DeleteExperience)

		// Messages (read-only for admin)
		admin.GET("/messages", adminContentH.ListMessages)
		admin.PATCH("/messages/:id/read", adminContentH.MarkMessageRead)
	}

	// ── Graceful Shutdown ──────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("🚀 DevFolio running on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("DevFolio stopped cleanly ✓")
}

func seedAdmin(adminRepo *repository.AdminRepository, cfg *config.Config) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return adminRepo.Seed(context.Background(), uuid.New().String(), cfg.AdminEmail, string(hash))
}
