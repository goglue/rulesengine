# Go tests
run-tests:
	go test ./... -bench=. -benchmem -race -covermode=atomic -coverprofile=coverage.out &&\
	go tool cover -func=coverage.out -o=coverage.out
