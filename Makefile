IMAGE_NAME  := auto-scaler
IMAGE_TAG   := latest
KUBE_DIR    := deploy/k8s

.PHONY: build deploy undeploy setup-monitoring teardown-monitoring load-test watch-hpa

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

# for minikube: loads image directly into the cluster without a registry
load-image: build
	minikube image load $(IMAGE_NAME):$(IMAGE_TAG)

deploy:
	kubectl apply -f $(KUBE_DIR)/deployment.yaml
	kubectl apply -f $(KUBE_DIR)/service.yaml
	kubectl apply -f $(KUBE_DIR)/ingress.yaml
	kubectl apply -f $(KUBE_DIR)/hpa.yaml

undeploy:
	kubectl delete -f $(KUBE_DIR)/deployment.yaml
	kubectl delete -f $(KUBE_DIR)/service.yaml
	kubectl delete -f $(KUBE_DIR)/ingress.yaml
	kubectl delete -f $(KUBE_DIR)/hpa.yaml

setup-monitoring:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update
	helm install kube-prometheus prometheus-community/kube-prometheus-stack \
		--set grafana.service.type=NodePort \
		--set prometheus.service.type=NodePort
	kubectl apply -f $(KUBE_DIR)/servicemonitor.yaml

teardown-monitoring:
	helm uninstall kube-prometheus
	kubectl delete -f $(KUBE_DIR)/servicemonitor.yaml

load-test:
	k6 run -e BASE_URL=$(BASE_URL) load-test/load-test.js

# watch HPA and pod count side by side
watch-hpa:
	watch -n 3 'kubectl get hpa auto-scaler-hpa && echo "" && kubectl get pods -l app=auto-scaler'
