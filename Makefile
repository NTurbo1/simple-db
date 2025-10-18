EXEC_NAME=simpleDB.bin

.PHONY: build run test clean

build:
	go build -o $(EXEC_NAME)

run: build
	./$(EXEC_NAME)

clean:
	rm -f $(EXEC_NAME)
