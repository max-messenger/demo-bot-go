package docs

//go:generate go tool swag init --parseInternal --parseDepth 1 --parseDependency -d ./.. -g internal/app/router/router.go -o ./
