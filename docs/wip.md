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
- PATCH /groups/{groupId}
- PATCH /groups/{groupId}/members/{userId}
- DELETE /groups/{groupId}/members/{userId}
- DELETE /groups/{groupId}
## Done:
- requireGroupMember
- requireGroupAdmin
- requireGroupOwner
- /groups/{groupId}/invite-code/regen
