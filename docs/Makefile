CTR_CMD :=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))
JEKYLL_VERSION := 3.8

vendor/bundle:
	mkdir -p vendor/bundle

build: vendor/bundle
	$(CTR_CMD) run --rm --volume="$(PWD):/srv/jekyll" --volume="$(PWD)/vendor/bundle:/usr/local/bundle" \
        -it docker.io/jekyll/jekyll:$(JEKYLL_VERSION) jekyll build

.PHONY: build

serve: vendor/bundle
	@echo ""
	@echo -e  "\033[1;37m open http://[::1]:4000 when the server is ready \033[0m"
	@echo ""
	$(CTR_CMD) run --rm --volume="$(PWD):/srv/jekyll"  --volume="$(PWD)/vendor/bundle:/usr/local/bundle" \
   		--publish [::1]:4000:4000  docker.io/jekyll/jekyll:$(JEKYLL_VERSION) jekyll serve

.PHONY: serve