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
  go get github.com/go-chi/httprate

## To Do:
- review codebase
- organize codebase (models, errors, constants)
- document API for frontend
- improve logging
- write tests?

## Done:
- added RequireEventOwner middleware
- finish implementing event endpoints
- changed event admin to event manager, and added event owner
