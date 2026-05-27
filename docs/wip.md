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
- implement event endpoints

- document API for frontend
- organize codebase (models, errors, constants)
- improve logging
- write tests?

## Done:
- create migration for events (events and event_participants)
- organized store functions into different files
- wrote db error translation helper
