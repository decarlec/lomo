OS := $(shell uname -s)

ifeq ($(OS),Linux)
install: install_linux
endif

ifeq ($(OS),Darwin)
install: install_mac
endif

db_bootstrap:
	cd bootstrap && go run bootstrap.go

run:
	go run . 

install_linux:
	$(MAKE) build_linux && sudo cp bin/lomo_linux /usr/local/bin/lomo

install_mac:
	$(MAKE) build_mac && sudo cp bin/lomo_mac /usr/local/bin/lomo

build_linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/lomo_linux main.go && chmod +x bin/lomo_linux

build_mac:
	CGO_ENABLED=1 GOODOS=darwin GOARCH=arm64 go build -o bin/lomo_mac main.go

build_win:
	GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/lomo_win.exe main.g
