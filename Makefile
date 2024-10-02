run: build
	@./bin/app

build: npm generate
	@go build -o bin/app .

generate:
	@templ generate

npm:
	@npm run build
