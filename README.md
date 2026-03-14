# ![Atoman 图标](web/public/favicon.ico) Atoman

![License](https://img.shields.io/badge/license-GPL-yellow) ![Vue 3](https://img.shields.io/badge/vue-3-42b883) ![Vite](https://img.shields.io/badge/vite-5-646cff)

Atoman希望构建一个开放、自由的言论空间，包括博客、论坛、内容订阅等板块。支持私有化部署。

## 特性
- 📰 Feed：占位页面，便于接入个性化订阅流。
- ✍️ Blog：基础博客入口，可扩展为文章列表与详情。
- 💬 Forum：社区讨论的起始页面，后续可接入话题、回复与通知。
- 🧰 技术栈：Vue 3 + Vite + TypeScript + Pinia + Naive UI + Tailwind，默认已接好路由与 UI 主题。

## 许可证
- 本项目使用 GPL 许可，简单来说，只要你使用本项目的任何源码，你都应该继承GPL协议。

---

## 子模块 / Sub-module: Atoman (Sub-project)

<div align="center">

**A minimalist archive for collecting all Atoman songs**

[![License: MIT](https://img.shields.io/badge/License-MIT-black.svg)](https://opensource.org/licenses/MIT)
[![Vue 3](https://img.shields.io/badge/Vue-3.x-black.svg)](https://vuejs.org/)
[![Go](https://img.shields.io/badge/Go-1.23-black.svg)](https://golang.org/)

[Live Demo](#) • [Documentation](#) • [Report Bug](#)

</div>

---

### Overview

Atoman is a black-and-white minimalist web application designed to collect, organize, and play all Atoman songs. Built with Vue 3 and Go, it features user authentication, song uploads, corrections, and a timeline-based UI inspired by archival aesthetics.

#### Features

- **Timeline View**: Chronological display of songs with year markers and visual timeline
- **Audio Player**: Full-featured player with play/pause, skip, shuffle, repeat, and volume control
- **User Authentication**: Register/login system with JWT tokens
- **Song Management**: Upload songs with metadata, album art, and lyrics
- **Correction System**: Submit corrections for song information
- **Admin Dashboard**: Review and approve pending uploads and corrections
- **Cloud Storage**: Media files stored on S3-compatible object storage
- **Dual Database Support**: SQLite for development, MySQL for production

---

### Tech Stack

#### Frontend
- **Vue 3** (Composition API + TypeScript)
- **Vite** - Build tool
- **Pinia** - State management
- **Vue Router** - Routing
- **Tailwind CSS** - Styling
- **Axios** - HTTP client

#### Backend
- **Go 1.23** with Gin framework
- **GORM** - ORM
- **JWT** - Authentication
- **AWS SDK** - S3-compatible storage
- **SQLite** / **MySQL** - Database

#### Infrastructure
- **Docker** - Containerization
- **Nginx** - Static file serving
- **Oracle Cloud** - Object storage

---

### Quick Start

#### Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local development)
- Node.js 18+ (for frontend development)

#### Development Environment (SQLite)

```bash
## Clone the repository
git clone https://github.com/yourusername/atoman.git
cd atoman

## Start development environment
docker-compose -f docker-compose.dev.yml up --build

## Access the application
## Frontend: http://localhost
## Backend API: http://localhost:8080
```

#### Production Environment (MySQL)

```bash
## Configure production environment
cp server/.env.production server/.env

## Edit server/.env with your MySQL credentials
## DATABASE_TYPE=mysql
## DATABASE_URL=user:password@tcp(host:port)/dbname?...

## Start production environment
docker-compose up --build

## Access the application
## Frontend: http://localhost
## Backend API: http://localhost:8080
```

---

### Configuration

#### Environment Variables

Create a `.env` file in the `server/` directory:

```env
## Environment
ENV=development
DATABASE_TYPE=sqlite
DATABASE_URL=./database.sqlite

## MySQL (Production)
DB_HOST=your-mysql-host
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your-password
DB_NAME=atoman

## Server
PORT=8080
GIN_MODE=debug

## JWT Secret
JWT_SECRET=your-secret-key

## S3 Storage
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
S3_BUCKET=your-bucket
S3_REGION=your-region
S3_ENDPOINT=https://your-endpoint
S3_URL_PREFIX=https://your-public-url
```

---

### API Documentation

#### Authentication

##### Register User
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "user",
  "email": "user@example.com",
  "password": "password123"
}
```

##### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### Songs

##### Get All Songs
```http
GET /api/songs
```

##### Get Song by ID
```http
GET /api/songs/:id
```

##### Create Song (Requires Authentication)
```http
POST /api/songs
Authorization: Bearer <token>
Content-Type: multipart/form-data

{
  "title": "Song Title",
  "artist": "Atoman",
  "album": "Album Name",
  "year": 2024,
  "track_number": 1,
  "lyrics": "Song lyrics...",
  "audio_file": <file>,
  "cover_file": <file>
}
```

#### Corrections

##### Submit Correction (Requires Authentication)
```http
POST /api/corrections
Authorization: Bearer <token>
Content-Type: application/json

{
  "song_id": 1,
  "field_name": "title",
  "current_value": "Wrong Title",
  "corrected_value": "Correct Title",
  "reason": "Official release has different title"
}
```

#### Admin

##### Get Pending Songs (Admin Only)
```http
GET /api/admin/pending
Authorization: Bearer <admin-token>
```

##### Approve Song (Admin Only)
```http
POST /api/admin/approve/:id
Authorization: Bearer <admin-token>
```

##### Reject Song (Admin Only)
```http
POST /api/admin/reject/:id
Authorization: Bearer <admin-token>
```

---

### Database Migration

To migrate data from SQLite to MySQL:

```bash
cd server

## Run migration tool
DB_HOST=your-host \
DB_PORT=3306 \
DB_USER=root \
DB_PASSWORD=your-password \
DB_NAME=atoman \
SQLITE_PATH=./database.sqlite \
go run cmd/migrate/main.go
```

---

### Project Structure

```
atoman/
├── web/                    # Frontend Vue 3 application
│   ├── src/
│   │   ├── views/         # Page components
│   │   ├── components/    # Reusable components
│   │   ├── stores/        # Pinia stores
│   │   ├── router/        # Vue Router config
│   │   └── assets/        # Static assets
│   ├── Dockerfile         # Frontend Docker config
│   └── package.json
├── server/                # Backend Go application
│   ├── main.go           # Entry point
│   ├── models.go         # Database models
│   ├── auth.go           # Authentication handlers
│   ├── songs.go          # Song handlers
│   ├── corrections.go    # Correction handlers
│   ├── admin.go          # Admin handlers
│   ├── storage.go        # S3 storage utilities
│   ├── middleware.go     # Middleware
│   ├── cmd/
│   │   ├── create_admin/ # Admin creation tool
│   │   └── migrate/      # Database migration tool
│   ├── Dockerfile        # Production Docker config
│   ├── Dockerfile.dev    # Development Docker config
│   └── go.mod
├── docker-compose.yml        # Production compose
├── docker-compose.dev.yml    # Development compose
└── README.md
```

---

### Design Philosophy

#### Minimalist Archive Aesthetic

- **Colors**: Pure black (#000000) and white (#FFFFFF) with gray accents
- **Typography**: Bold, black, ultra-tight tracking for headers; wide tracking for labels
- **Borders**: 2-4px solid black borders on all interactive elements
- **Shadows**: Hard drop shadows (e.g., `10px 10px 0px 0px rgba(0,0,0,1)`)
- **Images**: Grayscale filter applied to all cover art for visual consistency
- **Transitions**: Smooth 300-500ms transitions with hover state inversions

#### UI Components

- **Buttons**: Black background with white text, inverting on hover
- **Cards**: White background with black borders, hard shadow on hover
- **Inputs**: White with black border, focused state adds shadow
- **Timeline**: Vertical line with circular nodes, scales on active song

---

### Development

#### Frontend Development

```bash
cd web
npm install
npm run dev
```

#### Backend Development

```bash
cd server
go mod download
go run .
```

#### Creating Admin User

```bash
cd server
go run cmd/create_admin/main.go
```

---

### Deployment

#### Production with Cloudflare (Recommended)

For detailed Cloudflare setup with SSL/TLS, CDN, and optimization, see [CLOUDFLARE_DEPLOYMENT.md](CLOUDFLARE_DEPLOYMENT.md).

**Quick Start:**

1. **Point domain to server via Cloudflare DNS**
2. **Configure SSL in Cloudflare:**
   - Flexible SSL: No setup needed (current config)
   - Full SSL: Add Cloudflare Origin Certificate to `nginx/ssl/`
3. **Deploy:**
   ```bash
   docker-compose up -d
   ```

#### Using Docker Compose

```bash
## Production deployment
docker-compose up -d

## View logs
docker-compose logs -f

## Stop services
docker-compose down
```

#### Manual Deployment

##### Frontend
```bash
cd web
npm install
npm run build
## Serve dist/ folder with Nginx
```

##### Backend
```bash
cd server
CGO_ENABLED=1 go build -o main .
./main
```

---

### Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

### Acknowledgments

- Inspired by archival and museum collection interfaces
- Built with love for Atoman's discography
- Design philosophy: "Less is more, but black and white is everything"

---

<div align="center">

**[⬆ back to top](#atoman)**

Made with ⚫️ and ⚪️

</div>