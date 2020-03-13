.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/thisweek src/thisweek/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/getkaka src/getkaka/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/vote src/vote/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/location src/location/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/score src/score/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
