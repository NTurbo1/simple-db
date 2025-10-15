EXEC_NAME=simpleDB.bin

.PHONY: build run test

build:
	go build -o $(EXEC_NAME)

run: build
	./$(EXEC_NAME)
