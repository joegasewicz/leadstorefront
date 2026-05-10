IMG_NAME=josefdigital/platform:platform_web-v0.0.2
TEST_CONCURRENCY ?= 4

test:
	cd .. && go test -parallel $(TEST_CONCURRENCY) ./platform_web/routes

test_compile:
	cd .. && go test -c ./cmd/platform_web -o /tmp/leadstorefront-platform_web.test

docker_build:
	export DOCKER_DEFAULT_PLATFORM=linux/amd64 && docker build --no-cache --tag $(IMG_NAME) -f Dockerfile ..

docker_build_mac:
	docker build --no-cache --tag $(IMG_NAME) -f Dockerfile ..

docker_run:
	docker run --env-file ../infra/.env-dev --rm -p 8000:8000 $(IMG_NAME)

docker_push:
	docker push $(IMG_NAME)

docker_remove:
	docker rmi $(IMG_NAME)
