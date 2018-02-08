DEV_IMAGE = local/pspy-dev:latest
PROJECT_DIR = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

build-dev:
	docker build  -f docker/Dockerfile -t $(DEV_IMAGE) .

dev:
	docker run -it --rm -v $(PROJECT_DIR):/go/src/github.com/dominicbreuker/pspy $(DEV_IMAGE)
