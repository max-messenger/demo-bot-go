.PHONY: test_unit
test_unit:
	mkdir -p coverage
	go clean -testcache && go tool gotestsum -- -v -cover -coverprofile=coverage/unit.txt -race -run Unit ./...

.PHONY: test_cover
test_cover:
	grep -v -E -f .covignore coverage/unit.txt > coverage.filtered.txt
	go tool cover -func coverage.filtered.txt | fgrep total | awk '{print $$1, $$3}'

lint:
	golangci-lint run --fix

.PHONY: dev_env_up dev_env_down
dev_env_up: dev_env_down
	docker-compose pull
	docker-compose up -d --build
dev_env_down:
	docker-compose down -v --remove-orphans

.PHONY: generate
generate:
	go generate -x ./...
