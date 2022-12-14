# variables
name = tarjem
get_version = `git tag --sort=-version:refname | head -n 1`
inject_version = -ldflags "-X main.version=$(get_version)"
appname = $(name)-$(get_version)

# commands
install:
	go mod tidy

build:
	go build -v $(inject_version) ./...

compile:
	echo "Compiling for every OS and Platform"
	
	GOOS=freebsd GOARCH=amd64 go build $(inject_version) -o bin/$(appname)-freebsd-amd64 .
	GOOS=linux GOARCH=amd64 go build $(inject_version) -o bin/$(appname)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build $(inject_version) -o bin/$(appname)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 go build $(inject_version) -o bin/$(appname)-macos-amd64 .

clean:
	rm -f ./$(name) 
	rm -f -R ./bin

