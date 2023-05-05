tests = $(wildcard **/*.robot)
RF_VENV :=../_output/.venv
RF_BINARY :=../_output/.venv/bin/robot
OUTPUT_DIR ?= ../_output/e2e-$$(date +''%Y%m%d-%H%M%S)
RF_VARIABLES ?=suites/variables.yaml
RF_OPTIONS :=--randomize all --loglevel DEBUG -x junit.xml -V $(RF_VARIABLES) --outputdir $(OUTPUT_DIR)

.PHONY: setup dry-run all $(tests)

setup:
	python3 -m venv $(RF_VENV)  && \
	$(RF_VENV)/bin/python3 -m pip install -r requirements.txt

$(tests):
	$(RF_BINARY) \
		$(RF_OPTIONS) \
		$@

dry-run: ROBOT_OPTIONS=--dryrun
dry-run: $(tests)

all:
	$(RF_BINARY) \
		$(RF_OPTIONS) \
		./suites
