include .bingo/Variables.mk

JSONNET_VENDOR_DIR ?= grafana/vendor

JSONNETFMT_CMD := $(JSONNETFMT) -n 2 --max-blank-lines 2 --string-style s --comment-style s

.PHONY: jsonnet-format
jsonnet-format: $(JSONNETFMT)
	find . -name 'vendor' -prune -o -name '*.libsonnet' -print -o -name '*.jsonnet' -print | \
		xargs -n 1 -- $(JSONNETFMT_CMD) -i

.PHONY: jsonnet-lint
jsonnet-lint: $(JSONNET_LINT) ${JSONNET_VENDOR_DIR}
	find . -name 'vendor' -prune -o -name '*.libsonnet' -print -o -name '*.jsonnet' -print | \
		xargs -n 1 -- $(JSONNET_LINT) -J ${JSONNET_VENDOR_DIR}

.PHONY: dashboards
dashboards: $(JSONNET)
	-rm -rf grafana/dashboards/*
	$(JSONNET) -J ${JSONNET_VENDOR_DIR} -m grafana/dashboards grafana/dashboards.jsonnet
