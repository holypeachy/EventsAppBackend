# wip.md

## Dependencies:
> go get github.com/go-chi/chi/v5
  go get github.com/go-chi/cors
  go get github.com/jackc/pgx/v5
  go get github.com/golang-jwt/jwt/v5
  go get golang.org/x/crypto/bcrypt
  go get github.com/joho/godotenv
  go get github.com/pressly/goose/v3
  go get github.com/google/uuid

## To Do:
- POST /groups and GET /groups
- requireGroupMember
- requireGroupAdmin
- requireGroupOwner

## Done:
- implement /api/v1/auth/*
- implement auth middleware
- migration for groups and group_members
