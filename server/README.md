# Atoman Backend - Refactored Architecture

## Overview

The backend has been successfully refactored from a single monolithic `main.go` file into a modular architecture with separate components for different concerns.

## Architecture

### File Structure

```
server/
├── main.go          # Application entry point and route configuration
├── models.go        # Database models (User, Song, Correction)
├── middleware.go    # Authentication middleware
├── auth.go          # User authentication routes and handlers
├── songs.go         # Song management routes and handlers
├── corrections.go   # Correction management routes and handlers
├── storage.go       # S3/OCI Object Storage utilities
├── go.mod           # Go module dependencies
├── go.sum           # Go module checksums
├── .env             # Environment variables
├── schema.sql       # Database schema
├── start.sh         # Quick start script
└── TEST_RESULTS.md  # Connection test results
```

### Components

#### 1. `models.go`
- Contains all database model definitions
- `User` - User account model
- `Song` - Music track model
- `Correction` - User-submitted correction model

#### 2. `middleware.go`
- `AuthMiddleware()` - JWT token validation middleware

#### 3. `auth.go`
- User registration and login routes
- `SetupAuthRoutes()` - Configures auth endpoints
- `RegisterHandler()` - User registration logic
- `LoginHandler()` - User login logic

#### 4. `songs.go`
- Song CRUD operations
- `SetupSongRoutes()` - Configures song endpoints
- `GetSongsHandler()` - List all songs
- `GetSongHandler()` - Get single song
- `CreateSongHandler()` - Create song with optional file upload

#### 5. `corrections.go`
- Correction submission system
- `SetupCorrectionRoutes()` - Configures correction endpoints
- `CreateCorrectionHandler()` - Submit corrections

#### 6. `storage.go`
- S3/OCI Object Storage integration
- `InitS3Client()` - Initialize S3 client for OCI
- `ValidateS3Connection()` - Test S3 connectivity

#### 7. `main.go`
- Application bootstrap
- Database connection and migration
- Route setup
- Server startup

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login

### Songs
- `GET /api/songs` - List all songs
- `GET /api/songs/:id` - Get song details
- `POST /api/songs` - Create song (requires auth, supports file upload)

### Corrections
- `POST /api/corrections` - Submit correction (requires auth)

## Database Schema

Three main tables:
- `users` - User accounts
- `songs` - Music tracks
- `corrections` - User corrections

## Storage

- **Database**: MySQL 9.5.2-cloud
- **File Storage**: Oracle Cloud Infrastructure Object Storage (S3-compatible)
- **Authentication**: JWT tokens

## Running the Application

### Quick Start
```bash
cd server
./start.sh
```

### Manual Start
```bash
cd server
go run .
```

### Build and Run
```bash
cd server
go build -o atoman_server .
./atoman_server
```

## Testing

### Connection Tests
```bash
cd server
./run_tests.sh
```

### API Tests
```bash
cd server
./test_api.sh
```

### Manual Connection Test
If you need to run connection tests manually:
```bash
cd server
cp test/test_connections.go.backup test/test_connections.go
go run test/test_connections.go
rm test/test_connections.go
```

## Environment Variables

Required environment variables in `.env`:
- Database: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- S3/OCI: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `S3_BUCKET`, `S3_REGION`, `S3_NAMESPACE`, `S3_ENDPOINT`, `S3_URL_PREFIX`
- Auth: `JWT_SECRET`
- Server: `PORT`, `GIN_MODE`

## Benefits of Refactoring

1. **Modularity** - Each component has a single responsibility
2. **Maintainability** - Easier to modify and extend specific features
3. **Testability** - Individual components can be tested in isolation
4. **Readability** - Code is organized and easier to understand
5. **Reusability** - Components can be reused across different parts of the application

## Development Notes

- All handlers follow the same pattern: `SetupRoutes()` for configuration, individual handlers for logic
- Authentication is required for song creation and correction submission
- File uploads are handled via multipart form data
- OCI Object Storage uses `.compat.objectstorage` endpoint for S3 compatibility
- Database migrations are skipped if tables already exist (speeds up startup)

## Next Steps

1. ✅ Backend architecture refactored
2. ✅ All connections tested and verified
3. 🔄 Frontend integration
4. 🔄 Add more CRUD operations (update, delete)
5. 🔄 Implement correction approval workflow
6. 🔄 Add proper error handling and logging
7. 🔄 Add API documentation (Swagger/OpenAPI)
