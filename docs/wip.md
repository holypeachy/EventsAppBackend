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
- requireGroupMember
- requireGroupAdmin
- requireGroupOwner
- /groups/{groupId}/invite-code/regen

## Done:
- POST /groups
- GET /groups
- GET /groups/{id}
- POST /groups/join
- GET /groups/{id}/members
