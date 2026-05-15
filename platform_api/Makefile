IMG_NAME=josefdigital/platform:platform_api-v0.0.11
TEST_CONCURRENCY ?= 4

test:
	cd .. && go test -parallel $(TEST_CONCURRENCY) ./pkgs/... ./platform_api/routes

test_compile:
	cd .. && go test -c ./cmd/platform_api -o /tmp/leadstorefront-platform_api.test

docker_build:
	export DOCKER_DEFAULT_PLATFORM=linux/amd64 && docker build --no-cache --tag $(IMG_NAME) -f Dockerfile ..

docker_build_mac:
	docker build --no-cache --tag $(IMG_NAME) -f Dockerfile ..

docker_run:
	docker run --env-file ../infra/.env-dev --rm -p 8001:8001 $(IMG_NAME)

docker_push:
	docker push $(IMG_NAME)

docker_remove:
	docker rmi $(IMG_NAME)
