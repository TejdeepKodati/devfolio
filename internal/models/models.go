package models

import "time"

// ── Project ────────────────────────────────────────────────
type Project struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	LongDesc    string    `json:"long_description"`
	Tags        []string  `json:"tags"`
	GithubURL   string    `json:"github_url"`
	LiveURL     string    `json:"live_url"`
	ImageURL    string    `json:"image_url"`
	Featured    bool      `json:"featured"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpsertProjectRequest struct {
	Title       string   `json:"title"       binding:"required"`
	Description string   `json:"description" binding:"required"`
	LongDesc    string   `json:"long_description"`
	Tags        []string `json:"tags"`
	GithubURL   string   `json:"github_url"`
	LiveURL     string   `json:"live_url"`
	ImageURL    string   `json:"image_url"`
	Featured    bool     `json:"featured"`
	SortOrder   int      `json:"sort_order"`
}

// ── Skill ──────────────────────────────────────────────────
type Skill struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"` // e.g. "Backend", "Database", "Cloud"
	Proficiency int      `json:"proficiency"` // 1-100
	IconURL    string    `json:"icon_url"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}

type UpsertSkillRequest struct {
	Name        string `json:"name"        binding:"required"`
	Category    string `json:"category"    binding:"required"`
	Proficiency int    `json:"proficiency" binding:"required,min=1,max=100"`
	IconURL     string `json:"icon_url"`
	SortOrder   int    `json:"sort_order"`
}

// ── Experience ─────────────────────────────────────────────
type Experience struct {
	ID          string    `json:"id"`
	Company     string    `json:"company"`
	Role        string    `json:"role"`
	Description string    `json:"description"`
	StartDate   string    `json:"start_date"` // "Jan 2023"
	EndDate     string    `json:"end_date"`   // "Present" or "Dec 2024"
	IsCurrent   bool      `json:"is_current"`
	CompanyURL  string    `json:"company_url"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpsertExperienceRequest struct {
	Company     string `json:"company"     binding:"required"`
	Role        string `json:"role"        binding:"required"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"  binding:"required"`
	EndDate     string `json:"end_date"`
	IsCurrent   bool   `json:"is_current"`
	CompanyURL  string `json:"company_url"`
	SortOrder   int    `json:"sort_order"`
}

// ── Contact Message ────────────────────────────────────────
type Message struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type ContactRequest struct {
	Name    string `json:"name"    binding:"required,min=2"`
	Email   string `json:"email"   binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body"    binding:"required,min=10"`
}

// ── Auth ───────────────────────────────────────────────────
type AdminLoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ── Portfolio summary (single GET /api/portfolio) ─────────
type PortfolioData struct {
	Projects    []*Project    `json:"projects"`
	Skills      []*Skill      `json:"skills"`
	Experience  []*Experience `json:"experience"`
	FeaturedProjects []*Project `json:"featured_projects"`
}
