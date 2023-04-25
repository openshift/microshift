tests = $(wildcard **/*.robot)
RF_VENV :=../_output/.venv
RF_BINARY :=../_output/.venv/bin/robot
RF_OUT :=--outputdir ../_output/e2e-$$(date +''%Y%m%d-%H%M%S)

.PHONY: setup dry-run with-filter with-tag all $(tests)

setup:
	python3 -m venv $(RF_VENV)  && \
	$(RF_VENV)/bin/python3 -m pip install -r requirements.txt

$(tests):
	$(RF_BINARY) \
		$(RF_OPTIONS) \
		$(RF_OUT) \
		$@

dry-run: RF_OPTIONS=--dryrun
dry-run: $(tests)

all: RF_OPTIONS=-v USHIFT_HOST:$${USHIFT_HOST} -v USHIFT_USER:microshift -x xunit.xml
all: $(tests)
