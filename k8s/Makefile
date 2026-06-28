.PHONY: setup app k6-run port-forward clean argocd-install argocd-bootstrap argocd-password argocd-ui argocd-clean

NS = connectionpool-sandbox
ARGOCD_CHART_VERSION = 7.7.0

# --- GitOps (Argo CD) ---

argocd-install:
	helm repo add argo https://argoproj.github.io/argo-helm
	helm repo update
	kubectl apply -f argocd/namespace.yaml
	helm upgrade --install argocd argo/argo-cd \
		--namespace argocd \
		--version $(ARGOCD_CHART_VERSION) \
		--set "server.service.type=ClusterIP" \
		--wait

argocd-bootstrap: argocd-install
	kubectl apply -f argocd/root-application.yaml

argocd-password:
	@kubectl -n argocd get secret argocd-initial-admin-secret \
		-o jsonpath="{.data.password}" | base64 -d && echo

argocd-ui:
	kubectl port-forward svc/argocd-server -n argocd 8080:443 &
	@echo "Open https://localhost:8080  (admin / run 'make argocd-password')"

argocd-clean:
	helm uninstall argocd -n argocd || true
	kubectl delete namespace argocd || true

# --- Legacy imperative setup (Makefile-driven, kept for reference) ---

setup: helm-repos namespace mysql influxdb mysql-exporter tempo app prometheus grafana

helm-repos:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo update

namespace:
	kubectl apply -f namespace.yaml

mysql:
	kubectl apply -f mysql/ -n $(NS)
	kubectl wait --for=condition=ready pod/mysql-0 -n $(NS) --timeout=120s

influxdb:
	kubectl apply -f influxdb/ -n $(NS)

mysql-exporter:
	kubectl apply -f mysql-exporter/ -n $(NS)

app:
	minikube image build -t connectionpool-sandbox-app:latest ../app
	kubectl apply -f app/ -n $(NS)

tempo:
	kubectl apply -f tempo/ -n $(NS)

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
