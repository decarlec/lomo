db_bootstrap:
	cd bootstrap && go run bootstrap.go

run:
	go run . 

build_mac:
	CGO_ENABLED=1 GOODOS=darwin GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o bin/lomo_mac main.go

build_win:
	GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/lomo_win.exe main.go
