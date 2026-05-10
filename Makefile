TEST_CONCURRENCY ?= 4
PKG_PACKAGES := ./pkgs/...
API_TEST_PACKAGES := ./platform_api/routes
WEB_TEST_PACKAGES := ./platform_web/routes

test:
	go test -parallel $(TEST_CONCURRENCY) $(PKG_PACKAGES) $(API_TEST_PACKAGES) $(WEB_TEST_PACKAGES)

test_compile:
	go test -c ./cmd/platform_api -o /tmp/leadstorefront-platform_api.test
	go test -c ./cmd/platform_web -o /tmp/leadstorefront-platform_web.test

.PHONY: test test_compile
