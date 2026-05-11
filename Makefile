.PHONY: dev db/start db/stop db/logs

dev:
	go run cmd/api/main.go

db/start:
	docker compose up -d

db/stop:
	docker compose down

db/logs:
	docker compose logs -f postgres

migrate/up:
		migrate -path migrations -database "postgres://healthconnect:healthconnect_password@localhost:5432/healthconnect_db?sslmode=disable" up


migrate/down:
		migrate -path migrations -database "postgres://healthconnect:healthconnect_password@localhost:5432/healthconnect_db?sslmode=disable" down 1


migrate/create:
		migrate create -ext sql -dir migrations -seq $(name)
