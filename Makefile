.PHONY: all
all:
	go build -o gnAsteroid .

run: all
	./gnAsteroid -bind 127.0.0.1:8888 -asteroid-dir ./example
