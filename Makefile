BINARY = gnAsteroid

all:
	go build -C cmd -o ../${BINARY} .

install: all
	@if [ -w ~/bin/ ]; then cp ${BINARY} ~/bin; \
	else cp ${BINARY} /usr/local/bin; \
	fi

run: all
	# an example, visit on http://localhost:8888
	./gnAsteroid \
		-bind 127.0.0.1:8888 \
		-asteroid-dir ./example \
		-asteroid-name Xxx \
		-theme-dir themes/gnosmos.theme

.PHONY: all
