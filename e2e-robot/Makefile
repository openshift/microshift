
.PHONY: setup dry-run with-filter with-tag all

setup:
	python3 -m venv .venv && \
	.venv/bin/python3 -m pip install -r requirements.txt

dry-run:
	.venv/bin/robot \
		--dryrun \
		--outputdir ../_output/e2e-$$(date +'%Y%m%d-%H%M%S') \
		./tests/microshift.robot

with-filter:
	.venv/bin/robot \
		-v USHIFT_IP:$${USHIFT_IP} \
		-v USHIFT_USER:microshift \
		--test *router* \
		--outputdir ../_output/e2e-$$(date +'%Y%m%d-%H%M%S') \
		-x xunit.xml \
		./tests/microshift.robot

with-tag:
	.venv/bin/robot \
		-v USHIFT_IP:$${USHIFT_IP} \
		-v USHIFT_USER:microshift \
		--include smoke \
		--outputdir ../_output/e2e-$$(date +'%Y%m%d-%H%M%S') \
		-x xunit.xml \
		./tests/microshift.robot

all:
	.venv/bin/robot \
		-v USHIFT_IP:$${USHIFT_IP} \
		-v USHIFT_USER:microshift \
		--outputdir ../_output/e2e-$$(date +'%Y%m%d-%H%M%S') \
		-x xunit.xml \
		./tests/microshift.robot