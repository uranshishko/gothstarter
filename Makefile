run: npm generate build
	@./bin/app

build:
	@go build -o bin/app .

generate:
	@templ generate

npm:
	@npm run build
