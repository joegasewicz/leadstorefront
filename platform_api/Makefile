IMG_NAME=josefdigital/platform:platform_api-v0.0.2

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
