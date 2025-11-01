build:
	rm -rfd bin/
	mkdir bin/
	go build -o bin/send-jcode

install: build
	sudo cp bin/send-jcode /usr/local/bin