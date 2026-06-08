# DevFolio — Personal Portfolio

A full-stack portfolio website with a Go backend, JWT-protected admin panel,
PostgreSQL for content storage, and Redis for caching.

**Frontend:** Dark, editorial design with scroll animations.  
**Backend:** Go + Gin REST API + JWT admin auth.  
**Admin panel:** Create/edit/delete projects, skills, experience; read contact messages.

## Tech Stack
| Layer    | Technology                 |
|----------|----------------------------|
| Language | Go 1.21                    |
| Framework| Gin                        |
| Database | PostgreSQL 16              |
| Cache    | Redis 7                    |
| Auth     | JWT (HS256, 72h expiry)    |
| Frontend | Vanilla HTML/CSS/JS        |

## Project Structure
```
devfolio/
├── cmd/server/main.go              # Entrypoint + router + admin seed
├── internal/
│   ├── config/config.go
│   ├── db/db.go                    # PG pool + Redis client
│   ├── handlers/handlers.go        # Portfolio, Contact, AdminAuth, AdminContent
│   ├── middleware/auth.go          # JWT middleware
│   ├── models/models.go            # Project, Skill, Experience, Message
│   └── repository/repositories.go # All DB ops (upsert pattern)
├── web/static/
│   ├── index.html                  # Public portfolio (dark editorial design)
│   └── admin.html                  # JWT-protected CMS
├── migrations/001_initial.sql      # Schema + seed data (LinkLens & GoRelay pre-added)
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

## API Routes

### Public
| Method | Route          | Description                        |
|--------|----------------|------------------------------------|
| GET    | /              | Portfolio frontend                 |
| GET    | /admin         | Admin panel (requires JWT in browser)|
| GET    | /api/portfolio | All data in one shot (cached)      |
| GET    | /api/projects  | Projects list                      |
| GET    | /api/skills    | Skills list                        |
| GET    | /api/experience| Experience list                    |
| POST   | /api/contact   | Contact form submission            |
| POST   | /api/admin/login| Get admin JWT                     |

### Admin (JWT required — `Authorization: Bearer <token>`)
| Method | Route                         | Description             |
|--------|-------------------------------|-------------------------|
| POST   | /api/admin/projects/:id       | Upsert project          |
| DELETE | /api/admin/projects/:id       | Delete project          |
| POST   | /api/admin/skills/:id         | Upsert skill            |
| DELETE | /api/admin/skills/:id         | Delete skill            |
| POST   | /api/admin/experience/:id     | Upsert experience       |
| DELETE | /api/admin/experience/:id     | Delete experience       |
| GET    | /api/admin/messages           | List contact messages   |
| PATCH  | /api/admin/messages/:id/read  | Mark message as read    |

---

## Run Locally

### Docker (recommended)
```bash
cd devfolio
cp .env.example .env
docker compose up --build
# App:   http://localhost:8070
# Admin: http://localhost:8070/admin
```

Default admin credentials (from .env.example):
- Email: `admin@tejdeep.dev`
- Password: `changeme123`

### Manual
```bash
go mod tidy
createdb devfolio
psql devfolio -f migrations/001_initial.sql
cp .env.example .env
go run ./cmd/server
```

---

## Customise Your Content

**Option A — Admin panel (easiest)**
1. Open `http://localhost:8070/admin`
2. Login with your admin credentials
3. Add projects, skills, experience — they appear on the portfolio immediately

**Option B — Edit the seed SQL**  
Edit `migrations/001_initial.sql` — the `INSERT INTO projects/skills/experiences` section.

**Option C — REST API**
```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8070/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@tejdeep.dev","password":"changeme123"}' | jq -r .token)

# 2. Add a project
curl -X POST http://localhost:8070/api/admin/projects/new \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Project",
    "description": "What it does",
    "tags": ["Go","PostgreSQL"],
    "github_url": "https://github.com/tejdeep/myproject",
    "featured": true,
    "sort_order": 1
  }'
```

---

## Personalise the Frontend

Open `web/static/index.html` and change:
```html
<!-- Hero title -->
<h1 class="hero-title">Building systems<br>that <span>scale</span>.</h1>

<!-- Hero subtitle -->
<p class="hero-sub">I design and ship production-grade backend services...</p>

<!-- Social links -->
<a href="https://github.com/tejdeep" ...>
<a href="https://linkedin.com/in/tejdeep" ...>
<a href="mailto:tejdeep@example.com" ...>

<!-- Footer -->
<span>© 2025 Tejdeep. Built with <a href="https://go.dev">Go</a> + love.</span>
```

---

## Deploy to AWS (EC2 + RDS + ElastiCache)

```bash
# 1. Build
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o devfolio ./cmd/server

# 2. SCP to EC2
scp -i key.pem devfolio ubuntu@<IP>:~/
scp -i key.pem -r web ubuntu@<IP>:~/

# 3. On EC2 — run schema
psql $DATABASE_URL -f migrations/001_initial.sql

# 4. Systemd service
sudo tee /etc/systemd/system/devfolio.service > /dev/null <<EOF
[Unit]
Description=DevFolio Portfolio
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu
Environment=PORT=8070
Environment=DATABASE_URL=postgres://user:pass@<RDS>:5432/devfolio
Environment=REDIS_URL=redis://<ELASTICACHE>:6379
Environment=JWT_SECRET=your-production-secret-here
Environment=ADMIN_EMAIL=admin@tejdeep.dev
Environment=ADMIN_PASS=your-strong-password
Environment=ENV=production
ExecStart=/home/ubuntu/devfolio
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable --now devfolio

# 5. Nginx
sudo tee /etc/nginx/sites-available/devfolio > /dev/null <<EOF
server {
    listen 80;
    server_name tejdeep.dev www.tejdeep.dev;
    location / {
        proxy_pass http://localhost:8070;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/devfolio /etc/nginx/sites-enabled/
sudo nginx -s reload

# 6. Free HTTPS
sudo certbot --nginx -d tejdeep.dev -d www.tejdeep.dev
```

### Cheapest AWS setup (~$15–20/month)
| Service        | Tier               | Cost/month |
|----------------|-------------------|------------|
| EC2            | t3.micro           | ~$8        |
| RDS PostgreSQL | db.t3.micro (Free tier eligible) | Free 1yr / ~$15 after |
| ElastiCache    | cache.t3.micro     | ~$12       |
| Route 53       | Hosted zone        | ~$0.50     |

> **Tip:** For your portfolio, skip ElastiCache and use a local Redis on the same EC2.
> Redis on t3.micro handles portfolio traffic easily. Total cost drops to ~$8–10/month.
