default: test

include ../../../helpers/examples.mk

CHART := metricbeat
RELEASE := helm-metricbeat-upgrade
FROM := 7.10.0	# upgrade from version < 7.10.0 is failing due to selector
								# breaking change in https://github.com/elastic/helm-charts/pull/516

install:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm dependency update ../../
	../../../helpers/upgrade.sh --chart $(CHART) --release $(RELEASE) --from $(FROM)
	kubectl rollout status daemonset $(RELEASE)-metricbeat
	kubectl rollout status deployment $(RELEASE)-metricbeat-metrics
	kubectl rollout status deployment $(RELEASE)-kube-state-metrics

test: install goss

purge:
	helm del $(RELEASE)
