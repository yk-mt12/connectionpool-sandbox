.PHONY: setup app k6-run port-forward clean

NS = connectionpool-sandbox

setup: helm-repos namespace mysql influxdb mysql-exporter app prometheus grafana

helm-repos:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo update

namespace:
	kubectl apply -f namespace.yaml

mysql:
	helm upgrade --install mysql bitnami/mysql -n $(NS) -f helm-values/mysql.yaml --wait

influxdb:
	kubectl apply -f influxdb/ -n $(NS)

mysql-exporter:
	kubectl apply -f mysql-exporter/ -n $(NS)

app:
	minikube image build -t connectionpool-sandbox-app:latest ../app
	kubectl apply -f app/ -n $(NS)

prometheus:
	helm upgrade --install prometheus prometheus-community/prometheus -n $(NS) -f helm-values/prometheus.yaml

grafana:
	kubectl apply -f grafana-dashboard/ -n $(NS)
	helm upgrade --install grafana grafana/grafana -n $(NS) -f helm-values/grafana.yaml

k6-run:
	kubectl delete job k6-loadtest -n $(NS) --ignore-not-found
	kubectl apply -f k6/ -n $(NS)
	kubectl wait --for=condition=complete job/k6-loadtest -n $(NS) --timeout=300s
	kubectl logs job/k6-loadtest -n $(NS)

port-forward:
	kubectl port-forward svc/grafana 3000:80 -n $(NS) &
	kubectl port-forward svc/prometheus-server 9090:80 -n $(NS) &

clean:
	helm uninstall mysql -n $(NS) || true
	helm uninstall prometheus -n $(NS) || true
	helm uninstall grafana -n $(NS) || true
	kubectl delete namespace $(NS) || true
