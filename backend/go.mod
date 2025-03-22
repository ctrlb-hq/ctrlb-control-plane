module github.com/ctrlb-hq/ctrlb-control-plane/backend

go 1.23.1

require github.com/gorilla/mux v1.8.1

require github.com/mattn/go-sqlite3 v1.14.23

require (
	github.com/golang-jwt/jwt/v4 v4.5.1
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.31.0
)

require go.uber.org/multierr v1.10.0 // indirect
