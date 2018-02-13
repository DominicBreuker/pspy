DEV_IMAGE = local/pspy-dev:latest
PROJECT_DIR = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

build-dev:
	docker build  -f docker/Dockerfile -t $(DEV_IMAGE) .

dev:
	docker run -it --rm -v $(PROJECT_DIR):/go/src/github.com/dominicbreuker/pspy $(DEV_IMAGE)

release:
	docker run -it \
		       --rm \
		       -v $(PROJECT_DIR):/go/src/github.com/dominicbreuker/pspy \
			   --env CGO_ENABLED=0 \
			   --env GOOS=linux \
			   --env GOARCH=386 \
	           $(DEV_IMAGE) go build -a -ldflags '-extldflags "-static"' -o pspy/bin/pspy32 pspy/main.go
	docker run -it \
		       --rm \
		       -v $(PROJECT_DIR):/go/src/github.com/dominicbreuker/pspy \
			   --env CGO_ENABLED=0 \
			   --env GOOS=linux \
			   --env GOARCH=amd64 \
	           $(DEV_IMAGE) go build -a -ldflags '-extldflags "-static"' -o pspy/bin/pspy64 pspy/main.go
