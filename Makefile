.PHONY: all
all:
	go build -o website .

run: all
	./website -bind 127.0.0.1:8888 -asteroid-dir ./example
