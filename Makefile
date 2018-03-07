PROJECT_DIR = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

DEV_IMAGE      = local/pspy-development:latest
DEV_DOCKERFILE = $(PROJECT_DIR)/docker/Dockerfile.development

TEST_IMAGE      = local/pspy-testing:latest
TEST_DOCKERFILE = $(PROJECT_DIR)/docker/Dockerfile.testing

test:
	docker build -f $(TEST_DOCKERFILE) -t $(TEST_IMAGE) . 
	docker run -it --rm $(TEST_IMAGE)

dev-build:
	docker build -f $(DEV_DOCKERFILE) -t $(DEV_IMAGE) .

dev:
	docker run -it \
		       --rm \
			   -v $(PROJECT_DIR):/go/src/github.com/dominicbreuker/pspy \
			   -w "/go/src/github.com/dominicbreuker/pspy" \
			   $(DEV_IMAGE)

EXAMPLE_IMAGE      = local/pspy-example:latest
EXAMPLE_DOCKERFILE = $(PROJECT_DIR)/docker/Dockerfile.example

example:
	docker build -t $(EXAMPLE_IMAGE) -f $(EXAMPLE_DOCKERFILE) .
	docker run -it --rm $(EXAMPLE_IMAGE)

BUILD_IMAGE = golang:1.10-alpine

release:
	docker run -it \
		       --rm \
		       -v $(PROJECT_DIR):/go/src/github.com/dominicbreuker/pspy \
			   -w "/go/src/github.com/dominicbreuker" \
			   --env CGO_ENABLED=0 \
			   --env GOOS=linux \
			   --env GOARCH=386 \
	           $(BUILD_IMAGE) go build -a -ldflags '-extldflags "-static"' -o pspy/bin/pspy32 pspy/main.go
	docker run -it \
		       --rm \
		       -v $(PROJECT_DIR):/go/src/github.com/dominicbreuker/pspy \
			   -w "/go/src/github.com/dominicbreuker" \
			   --env CGO_ENABLED=0 \
			   --env GOOS=linux \
			   --env GOARCH=amd64 \
	           $(BUILD_IMAGE) go build -a -ldflags '-extldflags "-static"' -o pspy/bin/pspy64 pspy/main.go
