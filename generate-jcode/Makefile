build:
	rm -rfd bin/
	mkdir bin/
	go build -o bin/generate-jcode

install: build
	sudo cp bin/generate-jcode /usr/local/bin