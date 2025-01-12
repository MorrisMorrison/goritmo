.PHONY: clean build download-static test e2e build-prod

run-server:
	go run main.go

run-client:
	cd web-client/ && deno task dev --open

download-static:
	go run build/download_static.go

test:
	go test ./... -v

e2e:
	go run test/e2e.go

build:
	go build -o ./ytclipper main.go

build-prod: test download-static build
