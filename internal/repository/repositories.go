package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tejdeep/devfolio/internal/models"
)

var ErrNotFound = errors.New("record not found")

// ─────────────────────────────────────────────
//  Project Repository
// ─────────────────────────────────────────────

type ProjectRepository struct{ db *pgxpool.Pool }

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository { return &ProjectRepository{db: db} }

func (r *ProjectRepository) List(ctx context.Context) ([]*models.Project, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, description, long_description, tags, github_url, live_url, image_url,
		        featured, sort_order, created_at, updated_at
		 FROM projects ORDER BY sort_order ASC, created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProjects(rows)
}

func (r *ProjectRepository) ListFeatured(ctx context.Context) ([]*models.Project, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, description, long_description, tags, github_url, live_url, image_url,
		        featured, sort_order, created_at, updated_at
		 FROM projects WHERE featured = TRUE ORDER BY sort_order ASC LIMIT 6`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProjects(rows)
}

func (r *ProjectRepository) Upsert(ctx context.Context, id string, req *models.UpsertProjectRequest) (*models.Project, error) {
	tagsJSON, _ := json.Marshal(req.Tags)
	now := time.Now()
	var p models.Project
	var tagsStr string

	err := r.db.QueryRow(ctx,
		`INSERT INTO projects (id, title, description, long_description, tags, github_url, live_url, image_url, featured, sort_order, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$11)
		 ON CONFLICT (id) DO UPDATE SET
		   title = EXCLUDED.title, description = EXCLUDED.description,
		   long_description = EXCLUDED.long_description, tags = EXCLUDED.tags,
		   github_url = EXCLUDED.github_url, live_url = EXCLUDED.live_url,
		   image_url = EXCLUDED.image_url, featured = EXCLUDED.featured,
		   sort_order = EXCLUDED.sort_order, updated_at = EXCLUDED.updated_at
		 RETURNING id, title, description, long_description, tags::text, github_url, live_url, image_url, featured, sort_order, created_at, updated_at`,
		id, req.Title, req.Description, req.LongDesc, string(tagsJSON),
		req.GithubURL, req.LiveURL, req.ImageURL, req.Featured, req.SortOrder, now,
	).Scan(&p.ID, &p.Title, &p.Description, &p.LongDesc, &tagsStr,
		&p.GithubURL, &p.LiveURL, &p.ImageURL, &p.Featured, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(tagsStr), &p.Tags)
	return &p, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─────────────────────────────────────────────
//  Skill Repository
// ─────────────────────────────────────────────

type SkillRepository struct{ db *pgxpool.Pool }

func NewSkillRepository(db *pgxpool.Pool) *SkillRepository { return &SkillRepository{db: db} }

func (r *SkillRepository) List(ctx context.Context) ([]*models.Skill, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, category, proficiency, icon_url, sort_order, created_at
		 FROM skills ORDER BY category, sort_order ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []*models.Skill
	for rows.Next() {
		s := &models.Skill{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Category, &s.Proficiency, &s.IconURL, &s.SortOrder, &s.CreatedAt); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (r *SkillRepository) Upsert(ctx context.Context, id string, req *models.UpsertSkillRequest) (*models.Skill, error) {
	now := time.Now()
	s := &models.Skill{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO skills (id, name, category, proficiency, icon_url, sort_order, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (id) DO UPDATE SET
		   name=EXCLUDED.name, category=EXCLUDED.category, proficiency=EXCLUDED.proficiency,
		   icon_url=EXCLUDED.icon_url, sort_order=EXCLUDED.sort_order
		 RETURNING id, name, category, proficiency, icon_url, sort_order, created_at`,
		id, req.Name, req.Category, req.Proficiency, req.IconURL, req.SortOrder, now,
	).Scan(&s.ID, &s.Name, &s.Category, &s.Proficiency, &s.IconURL, &s.SortOrder, &s.CreatedAt)
	return s, err
}

func (r *SkillRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM skills WHERE id = $1`, id)
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// ─────────────────────────────────────────────
//  Experience Repository
// ─────────────────────────────────────────────

type ExperienceRepository struct{ db *pgxpool.Pool }

func NewExperienceRepository(db *pgxpool.Pool) *ExperienceRepository {
	return &ExperienceRepository{db: db}
}

func (r *ExperienceRepository) List(ctx context.Context) ([]*models.Experience, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, company, role, description, start_date, end_date, is_current, company_url, sort_order, created_at
		 FROM experiences ORDER BY sort_order ASC, created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exps []*models.Experience
	for rows.Next() {
		e := &models.Experience{}
		if err := rows.Scan(&e.ID, &e.Company, &e.Role, &e.Description, &e.StartDate, &e.EndDate,
			&e.IsCurrent, &e.CompanyURL, &e.SortOrder, &e.CreatedAt); err != nil {
			return nil, err
		}
		exps = append(exps, e)
	}
	return exps, nil
}

func (r *ExperienceRepository) Upsert(ctx context.Context, id string, req *models.UpsertExperienceRequest) (*models.Experience, error) {
	now := time.Now()
	e := &models.Experience{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO experiences (id, company, role, description, start_date, end_date, is_current, company_url, sort_order, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (id) DO UPDATE SET
		   company=EXCLUDED.company, role=EXCLUDED.role, description=EXCLUDED.description,
		   start_date=EXCLUDED.start_date, end_date=EXCLUDED.end_date, is_current=EXCLUDED.is_current,
		   company_url=EXCLUDED.company_url, sort_order=EXCLUDED.sort_order
		 RETURNING id, company, role, description, start_date, end_date, is_current, company_url, sort_order, created_at`,
		id, req.Company, req.Role, req.Description, req.StartDate, req.EndDate,
		req.IsCurrent, req.CompanyURL, req.SortOrder, now,
	).Scan(&e.ID, &e.Company, &e.Role, &e.Description, &e.StartDate, &e.EndDate,
		&e.IsCurrent, &e.CompanyURL, &e.SortOrder, &e.CreatedAt)
	return e, err
}

func (r *ExperienceRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM experiences WHERE id = $1`, id)
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// ─────────────────────────────────────────────
//  Message Repository
// ─────────────────────────────────────────────

type MessageRepository struct{ db *pgxpool.Pool }

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository { return &MessageRepository{db: db} }

func (r *MessageRepository) Create(ctx context.Context, m *models.Message) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO messages (id, name, email, subject, body, is_read, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		m.ID, m.Name, m.Email, m.Subject, m.Body, false, m.CreatedAt,
	)
	return err
}

func (r *MessageRepository) List(ctx context.Context) ([]*models.Message, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, email, subject, body, is_read, created_at
		 FROM messages ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*models.Message
	for rows.Next() {
		m := &models.Message{}
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Subject, &m.Body, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *MessageRepository) MarkRead(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE messages SET is_read = TRUE WHERE id = $1`, id)
	return err
}

// ── Admin user (simple single-admin model) ────────────────

type AdminRepository struct{ db *pgxpool.Pool }

func NewAdminRepository(db *pgxpool.Pool) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) GetByEmail(ctx context.Context, email string) (id, hash string, err error) {
	err = r.db.QueryRow(ctx,
		`SELECT id, password_hash FROM admins WHERE email = $1`, email,
	).Scan(&id, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return
}

func (r *AdminRepository) Seed(ctx context.Context, id, email, hash string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO admins (id, email, password_hash) VALUES ($1,$2,$3) ON CONFLICT (email) DO NOTHING`,
		id, email, hash,
	)
	return err
}

// ── helpers ────────────────────────────────────────────────

func scanProjects(rows pgx.Rows) ([]*models.Project, error) {
	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		var tagsStr string
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.LongDesc, &tagsStr,
			&p.GithubURL, &p.LiveURL, &p.ImageURL, &p.Featured, &p.SortOrder,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		json.Unmarshal([]byte(tagsStr), &p.Tags)
		projects = append(projects, p)
	}
	return projects, nil
}
