BINARY = gnAsteroid

all:
	go build -C cmd -o ../${BINARY} .

install: all
	@if [ -w ~/bin/ ]; then cp ${BINARY} ~/bin; \
	else cp ${BINARY} /usr/local/bin; \
	fi

run: all
	./gnAsteroid -bind 127.0.0.1:8888 -asteroid-dir ./example


.PHONY: all
