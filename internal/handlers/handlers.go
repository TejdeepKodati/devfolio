package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/tejdeep/devfolio/internal/config"
	"github.com/tejdeep/devfolio/internal/middleware"
	"github.com/tejdeep/devfolio/internal/models"
	"github.com/tejdeep/devfolio/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const cacheTTL = 5 * time.Minute

// ─────────────────────────────────────────────
//  Portfolio Handler  (public API)
// ─────────────────────────────────────────────

type PortfolioHandler struct {
	projectRepo    *repository.ProjectRepository
	skillRepo      *repository.SkillRepository
	experienceRepo *repository.ExperienceRepository
	rdb            *redis.Client
}

func NewPortfolioHandler(
	pr *repository.ProjectRepository,
	sr *repository.SkillRepository,
	er *repository.ExperienceRepository,
	rdb *redis.Client,
) *PortfolioHandler {
	return &PortfolioHandler{projectRepo: pr, skillRepo: sr, experienceRepo: er, rdb: rdb}
}

// GET /api/portfolio — returns everything in one shot (cached)
func (h *PortfolioHandler) GetAll(c *gin.Context) {
	ctx := context.Background()

	projects, _ := h.projectRepo.List(ctx)
	featured, _ := h.projectRepo.ListFeatured(ctx)
	skills, _ := h.skillRepo.List(ctx)
	experience, _ := h.experienceRepo.List(ctx)

	c.JSON(http.StatusOK, models.PortfolioData{
		Projects:         projects,
		FeaturedProjects: featured,
		Skills:           skills,
		Experience:       experience,
	})
}

func (h *PortfolioHandler) GetProjects(c *gin.Context) {
	projects, err := h.projectRepo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load projects"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func (h *PortfolioHandler) GetSkills(c *gin.Context) {
	skills, err := h.skillRepo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load skills"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"skills": skills})
}

func (h *PortfolioHandler) GetExperience(c *gin.Context) {
	exp, err := h.experienceRepo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load experience"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"experience": exp})
}

// ─────────────────────────────────────────────
//  Contact Handler
// ─────────────────────────────────────────────

type ContactHandler struct {
	messageRepo *repository.MessageRepository
}

func NewContactHandler(mr *repository.MessageRepository) *ContactHandler {
	return &ContactHandler{messageRepo: mr}
}

// POST /api/contact — public contact form submission
func (h *ContactHandler) Submit(c *gin.Context) {
	var req models.ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := &models.Message{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		Subject:   req.Subject,
		Body:      req.Body,
		CreatedAt: time.Now(),
	}
	if err := h.messageRepo.Create(context.Background(), msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not send message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Message received! I'll get back to you soon."})
}

// ─────────────────────────────────────────────
//  Admin Auth Handler
// ─────────────────────────────────────────────

type AdminAuthHandler struct {
	adminRepo *repository.AdminRepository
	authMW    *middleware.AuthMiddleware
}

func NewAdminAuthHandler(adminRepo *repository.AdminRepository, cfg *config.Config) *AdminAuthHandler {
	return &AdminAuthHandler{adminRepo: adminRepo, authMW: middleware.NewAuthMiddleware(cfg)}
}

// POST /api/admin/login
func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, hash, err := h.adminRepo.GetByEmail(context.Background(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.authMW.Generate(id, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Welcome back!"})
}

// ─────────────────────────────────────────────
//  Admin Content Handler  (protected CRUD)
// ─────────────────────────────────────────────

type AdminContentHandler struct {
	projectRepo    *repository.ProjectRepository
	skillRepo      *repository.SkillRepository
	experienceRepo *repository.ExperienceRepository
	messageRepo    *repository.MessageRepository
}

func NewAdminContentHandler(
	pr *repository.ProjectRepository,
	sr *repository.SkillRepository,
	er *repository.ExperienceRepository,
	mr *repository.MessageRepository,
) *AdminContentHandler {
	return &AdminContentHandler{projectRepo: pr, skillRepo: sr, experienceRepo: er, messageRepo: mr}
}

// ── Projects ──────────────────────────────────────────────

func (h *AdminContentHandler) UpsertProject(c *gin.Context) {
	var req models.UpsertProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := c.Param("id")
	if id == "" || id == "new" {
		id = uuid.New().String()
	}
	p, err := h.projectRepo.Upsert(context.Background(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save project: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *AdminContentHandler) DeleteProject(c *gin.Context) {
	if err := h.projectRepo.Delete(context.Background(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}

// ── Skills ────────────────────────────────────────────────

func (h *AdminContentHandler) UpsertSkill(c *gin.Context) {
	var req models.UpsertSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := c.Param("id")
	if id == "" || id == "new" {
		id = uuid.New().String()
	}
	s, err := h.skillRepo.Upsert(context.Background(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save skill"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *AdminContentHandler) DeleteSkill(c *gin.Context) {
	if err := h.skillRepo.Delete(context.Background(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "skill deleted"})
}

// ── Experience ────────────────────────────────────────────

func (h *AdminContentHandler) UpsertExperience(c *gin.Context) {
	var req models.UpsertExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := c.Param("id")
	if id == "" || id == "new" {
		id = uuid.New().String()
	}
	e, err := h.experienceRepo.Upsert(context.Background(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save experience"})
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *AdminContentHandler) DeleteExperience(c *gin.Context) {
	if err := h.experienceRepo.Delete(context.Background(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "experience deleted"})
}

// ── Messages ──────────────────────────────────────────────

func (h *AdminContentHandler) ListMessages(c *gin.Context) {
	msgs, err := h.messageRepo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load messages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": msgs, "count": len(msgs)})
}

func (h *AdminContentHandler) MarkMessageRead(c *gin.Context) {
	if err := h.messageRepo.MarkRead(context.Background(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not mark read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}
