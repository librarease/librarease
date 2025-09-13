# GitHub Copilot Instructions for librarease

## Project Overview

**librarease** is a library management system built with Go and PostgreSQL, following a clean architecture pattern with distinct layers: handlers, usecases, and database repositories.

## Architecture & Key Patterns

### Layer Structure
- **Handlers** (`internal/server/*_handlers.go`): HTTP API endpoints using Echo framework
- **Usecases** (`internal/usecase/`): Business logic layer with interface-based dependency injection
- **Database** (`internal/database/`): GORM-based repository implementations
- **Configuration** (`internal/config/`): Environment constants and context keys

### Request/Response Pattern
All handlers follow this pattern:
```go
type SomeRequest struct {
    Field string `json:"field" validate:"required"`
}

func (s *Server) SomeHandler(ctx echo.Context) error {
    var req SomeRequest
    if err := ctx.Bind(&req); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    if err := s.validator.Struct(req); err != nil {
        return ctx.JSON(422, map[string]string{"error": err.Error()})
    }
    // Call usecase...
}
```

### Database Models
- All models use `uuid.UUID` primary keys with `uuid_generate_v4()` default
- GORM soft deletes with `DeletedAt *gorm.DeletedAt`
- Table names explicitly defined with `TableName()` method
- Timestamps: `CreatedAt time.Time`, `UpdatedAt time.Time`

### Service Interface Pattern
The `server.Service` interface acts as the main contract between HTTP handlers and business logic, implemented by the database service which wraps usecases.

### Dependency Injection
Usecases receive interfaces for:
- `Repository`: Database operations
- `IdentityProvider`: Firebase auth
- `FileStorageProvider`: MinIO/S3
- `Mailer`: SMTP email
- `Dispatcher`: Push notifications

## Development Workflow

### Essential Commands
```bash
make docker-run    # Start PostgreSQL + Redis containers
make run          # Run the API server locally
make test         # Unit tests for server + usecase layers
make itest        # Integration tests for database layer
make watch        # Live reload with air (if installed)
```

### Environment Setup
The app uses `.env` files loaded via `godotenv`. Key environment variables are defined as constants in `internal/config/config.go`:
- Database: `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USER`, `DB_PASSWORD`
- MinIO: `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`
- Firebase: Service account JSON file path
- SMTP: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`

### Context Conventions
Use context keys from `internal/config/config.go`:
- `CTX_KEY_USER_ID`: Extract current user ID
- `CTX_KEY_USER_ROLE`: Extract user role
- Headers: `X-Uid` for user identification, `X-Client-Id` for client tracking

## Integration Points

### Real-time Features
- WebSocket endpoint at `/ws` for live notifications
- PostgreSQL `LISTEN/NOTIFY` for database events
- Notification hub in `database/notification.go`

### File Storage
- MinIO-compatible S3 storage via `internal/filestorage/`
- Pre-signed URLs with 15-minute expiration
- Temp and public path separation

### Push Notifications
- Multi-platform push via `internal/push/dispatcher.go`
- APNS for iOS, WebPush for web clients
- Token management in `push_token` table

## Testing Strategy

- **Unit tests**: `internal/server/` and `internal/usecase/` 
- **Integration tests**: `internal/database/` with real PostgreSQL
- Use `make itest` for database integration tests specifically
- Test files follow `*_test.go` convention

## Common Patterns to Follow

1. **Error handling**: Return structured JSON errors with appropriate HTTP status codes
2. **Validation**: Use struct tags with `github.com/go-playground/validator/v10`
3. **Database queries**: Use GORM with context, include soft delete considerations
4. **File operations**: Use the usecase layer for file storage abstraction
5. **Authentication**: Middleware extracts user context, handlers use context keys
6. **Migrations**: GORM AutoMigrate + custom SQL in `database/migrations/`

When adding new features, maintain the clean architecture boundaries and follow the established request/response patterns.