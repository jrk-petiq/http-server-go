run:
	go build -o out && ./out

runserver:
	@echo 'Starting server...'
	go run main.go
