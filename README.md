# WSO-scaling-and-ha-in-stateless-apps
co używamy:

wget https://github.com/bojand/ghz/releases/latest/download/ghz-linux-x86_64.tar.gz
tar -xzf ghz-linux-x86_64.tar.gz
sudo mv ghz /usr/local/bin/
ghz --version

go 1.26

sudo apt install -y protobuf-compiler

curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

PROTOC_VERSION=34.1
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip
unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d $HOME/.local
rm protoc-${PROTOC_VERSION}-linux-x86_64.zip
export PATH="$HOME/.local/bin:$PATH"

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest


curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
minikube start --cpus=4 --memory=4096 --driver=docker
minikube addons enable metrics-server   # wymagany przez HPA
minikube addons enable ingress          # dla load balancera NGINX
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install monitoring prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace
kubectl get pods -n monitoring
minikube dashboard

drugi terminal:

kubectl port-forward deployment/monitoring-grafana -n monitoring 3000:3000
i w przeglądarce localhost:3000
(kubectl --namespace monitoring get secrets monitoring-grafana -o jsonpath="{.data.admin-password}" | base64 -d ; echo) - hasło do admina
trzeci terminal:

eval $(minikube docker-env)
docker build -t grpc-app:latest .

kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f hpa.yaml
kubectl apply -f ingress.yaml

test
kubectl get pods -w
kubectl get hpa grpc-app-hpa
kubectl port-forward service/grpc-app-service 50051:50051


czwarty terminal:
sudo sh -c "echo '$(minikube ip) grpc-app.local' >> /etc/hosts"

ghz --insecure --call dot_product.DotProductService.Calculate --concurrency 100 --rps 10000 --duration 1200s --data-file ./data.json grpc-app.local:80

kontrolowana awaria:
minikube ssh "curl http://$(kubectl get pod grpc-app-545cd7b465-nw2xn  -o jsonpath='{.status.podIP}'):8081/panic"

ograniczenie zasobów:
kubectl apply -f  manifests/quota.yaml
