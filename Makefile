# Paths
ROOT_DIR := $(patsubst %/,%, $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
DOCKER_DIR := $(ROOT_DIR)/docker
DOCKER_COMPOSE_FILE := docker-compose.yaml

# Colorized Prints
BOLD := $(shell tput bold)
RED := $(shell tput setaf 1)
BLUE := $(shell tput setaf 4)
RESET := $(shell tput sgr0)

# Targets

.PHONY: all
all: docker-up

###@ docker-up: Start docker services

.PHONY: docker-up
docker-up:
	@-echo "$(BLUE)Starting Docker services.$(RESET)"
	docker-compose -f $(DOCKER_DIR)/$(DOCKER_COMPOSE_FILE) up

###@ docker-down: Stop and remove docker services

.PHONY: docker-down
docker-down:
	@-echo "$(BLUE)Stopping Docker services.$(RESET)"
	docker-compose -f $(DOCKER_DIR)/$(DOCKER_COMPOSE_FILE) down

###@ help: Help

.PHONY: help
help: Makefile
	@-echo "Usage:\n  make $(BLUE)<target>$(RESET)"
	@-echo
	@-sed -n 's/^###@//p' $< | column -t -s ':' | sed -e 's/[^ ]*/ &/2'
