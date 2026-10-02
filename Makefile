export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: all build run test clean mock proto

all: build

up: BUILD_FLAG=
up: _launch

start: BUILD_FLAG=--build
start: _launch

_launch:
	@echo "Checking/creating Docker network (vpn_network)..."
	docker network create vpn_network 2>/dev/null || true
	@echo "Pulling latest docker images..."
	docker compose -f docker-compose.group_jupiter.yml pull || true
	docker compose -f docker-compose.group_saturn.yml pull || true
	@echo "Starting Group Jupiter (Cluster A)..."
	docker compose -f docker-compose.group_jupiter.yml up -d $(BUILD_FLAG)
	@echo "Starting Group Saturn Core Infrastructure (Cluster B)..."
	docker compose -f docker-compose.group_saturn.yml up -d $(BUILD_FLAG) group_saturn vault vault_init rabbitmq redis minio createbuckets
	@echo "Initializing databases..."
	bash scripts/init-dbs.sh
	@echo "Applying migrations..."
	bash scripts/migrate.sh
	@echo "Checking OpenStreetMap data & OSRM graphs..."
	bash scripts/setup-osm.sh
	@echo "Starting remaining Group Saturn microservices (Cluster B)..."
	BUILD_FLAG="$(BUILD_FLAG)" bash scripts/start-group-saturn.sh
	@echo "Project started successfully!"
	@echo "=================================================="
	@echo "🔗 ТУННЕЛЬ CLOUDFLARE АКТИВЕН:"
	@echo "Backend API: https://api.yolyolakay.com"
	@echo "=================================================="

down:
	@echo "Stopping Group Jupiter..."
	docker compose -f docker-compose.group_jupiter.yml down --remove-orphans
	@echo "Stopping Group Saturn..."
	docker compose -f docker-compose.group_saturn.yml down --remove-orphans
	@echo "All services stopped."

clear:
	@echo "Removing all data volumes and networks..."
	docker compose -f docker-compose.group_jupiter.yml down -v --remove-orphans
	docker compose -f docker-compose.group_saturn.yml down -v --remove-orphans
	docker network rm vpn_network 2>/dev/null || true
	@echo "All data cleared successfully!"

build:
	@echo "Building all services into bin/..."
	mkdir -p bin
	@for d in cmd/*; do \
		if [ -d "$$d" ]; then \
			name=$$(basename "$$d"); \
			echo "Building $$name..."; \
			go build -o bin/$$name ./$$d; \
		fi \
	done

run:
	@echo "Run specific service using 'go run cmd/<service>/main.go'"

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning up..."
	rm -rf bin/ *_service *_gateway

restore-db:
	@echo "Restoring database dumps..."
	@for db in racoon burunov spidy ford kibertron gobliddin; do \
		if [ -f "deployments/db_dumps/$$db.sql" ]; then \
			echo "Restoring $$db in Group Jupiter..."; \
			cat deployments/db_dumps/$$db.sql | docker exec -i group_jupiter psql -U KernelStingray -d $$db; \
		fi; \
	done
	@for db in romanova pascal girard senna shennon caprio; do \
		if [ -f "deployments/db_dumps/$$db.sql" ]; then \
			echo "Restoring $$db in Group Saturn..."; \
			cat deployments/db_dumps/$$db.sql | docker exec -i group_saturn psql -U KernelStingray -d $$db; \
		fi; \
	done
	@echo "Databases restored!"

mock:
	@echo "Generating mocks..."
	go run go.uber.org/mock/mockgen@latest -source=internal/auth/domain/otp.go -destination=pkg/mocks/auth/otp_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/auth/domain/session.go -destination=pkg/mocks/auth/session_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/ride/domain/repository.go -destination=pkg/mocks/ride/ride_repo_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/ride/domain/cache.go -destination=pkg/mocks/ride/ride_cache_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/ride/domain/usecase.go -destination=pkg/mocks/ride/ride_usecase_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/billing/domain/repository.go -destination=pkg/mocks/billing/billing_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/billing/domain/usecase.go -destination=pkg/mocks/billing/billing_usecase_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/user/domain/user.go -destination=pkg/mocks/user/user_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/user/domain/usecase.go -destination=pkg/mocks/user/usecase_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/user/domain/car.go -destination=pkg/mocks/user/car_mock.go -package=mocks
	go run go.uber.org/mock/mockgen@latest -source=internal/notification/domain/repository.go -destination=pkg/mocks/notification/notification_mock.go -package=mocks


proto:
	@echo "Generating Protobuf & gRPC code into pkg/api/grpc/..."
	rm -rf pkg/api/grpc/user pkg/api/grpc/ride pkg/api/grpc/billing
	mkdir -p pkg/api/grpc/user pkg/api/grpc/ride pkg/api/grpc/billing
	./protoc_bin/bin/protoc --proto_path=api/proto \
		--proto_path=protoc_bin/include \
		--go_out=pkg/api/grpc --go_opt=paths=source_relative \
		--go-grpc_out=pkg/api/grpc --go-grpc_opt=paths=source_relative \
		api/proto/user/user.proto
	./protoc_bin/bin/protoc --proto_path=api/proto \
		--proto_path=protoc_bin/include \
		--go_out=pkg/api/grpc --go_opt=paths=source_relative \
		--go-grpc_out=pkg/api/grpc --go-grpc_opt=paths=source_relative \
		api/proto/ride/ride.proto
	./protoc_bin/bin/protoc --proto_path=api/proto \
		--proto_path=protoc_bin/include \
		--go_out=pkg/api/grpc --go_opt=paths=source_relative \
		--go-grpc_out=pkg/api/grpc --go-grpc_opt=paths=source_relative \
		api/proto/billing/billing.proto

swagger:
	@echo "Generating unified Swagger API documentation..."
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d ./cmd/api_gateway,./internal/analytics/delivery/http,./internal/antifraud/delivery/http,./internal/auth/delivery/http,./internal/backoffice/delivery/http,./internal/billing/delivery/http,./internal/helpdesk/delivery/http,./internal/media/delivery/http,./internal/notification/delivery/http,./internal/parcel/delivery/http,./internal/rating/delivery/http,./internal/ride/delivery/http,./internal/routing/delivery/http,./internal/telematics/delivery/http,./internal/user/delivery/http -o docs --parseDependency --parseInternal


