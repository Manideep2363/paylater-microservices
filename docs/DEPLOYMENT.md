# PayLater Deployment

The Kubernetes manifests are split so local Docker Desktop uses ingress-nginx and AWS K3s uses its default Traefik controller. Both environments use the same same-origin frontend API path: `/api`.

The resource values below are initial values for a small single-node cluster. Tune them after observing actual usage.

## Local Deployment

Requirements:

- Docker Desktop Kubernetes enabled
- `kubectl` using the `docker-desktop` context
- `paylater.local` resolving to `127.0.0.1` in the hosts file

From the `paylater-microservices` directory:

```powershell
kubectl config use-context docker-desktop
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
kubectl get ingressclass nginx
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/config/configmap.yaml
```

Run PowerShell as Administrator once to add the local hostname:

```powershell
Add-Content "$env:SystemRoot\System32\drivers\etc\hosts" "127.0.0.1 paylater.local"
```

For local use, create the Secret without committing it. The `secret.example.yaml` file documents the required keys but should not be applied with its placeholders:

```powershell
kubectl create secret generic paylater-secret `
  --namespace paylater `
  --from-literal=MYSQL_ROOT_PASSWORD='your-local-password' `
  --from-literal=DB_PASSWORD='your-local-password' `
  --from-literal=JWT_SECRET='your-local-jwt-secret' `
  --from-literal=INTERNAL_API_TOKEN='your-local-internal-token' `
  --from-literal=ADMIN_EMAIL='admin@example.com' `
  --from-literal=ADMIN_PASSWORD='your-local-admin-password'
```

Deploy in dependency order:

```powershell
kubectl apply -f k8s/database/mysql.yaml
kubectl apply -f k8s/backend/user/user-service.yaml
kubectl apply -f k8s/backend/merchant/merchant-service.yaml
kubectl apply -f k8s/backend/auth/auth-service.yaml
kubectl apply -f k8s/backend/ledger/ledger-service.yaml
kubectl apply -f k8s/backend/report/report-service.yaml
kubectl apply -f k8s/backend/gateway/api-gateway.yaml
kubectl apply -f k8s/frontend/frontend-deployment.yaml
kubectl apply -f k8s/frontend/frontend-service.yaml
kubectl apply -f k8s/ingress/local-ingress.yaml
```

Verify:

```powershell
kubectl get pods,svc,ingress -n paylater
```

Open `http://paylater.local`. The local Ingress preserves the existing nginx regex rewrite: `/api/register` becomes `/register` at the gateway, while `/` goes to the frontend.

## AWS Deployment

Use a Linux EC2 instance with an attached security group allowing TCP 22 from your administrator IP and TCP 80 from intended clients. For this complete stack, `t3.small` or larger is the recommended minimum; a 1 GiB micro instance is suitable only for a best-effort demonstration and may run out of memory. Confirm your account's current Free Tier eligibility and instance options in the AWS console.

Install K3s on the EC2 instance:

```bash
curl -sfL https://get.k3s.io | sh -
sudo kubectl get nodes
```

Clone the repository and enter the deployment tree:

```bash
git clone <your-repository-url>
cd <clone-directory>/paylater-microservices
```

Create the namespace, configuration, and real Secret. Do not apply `secret.example.yaml` as the production Secret:

```bash
sudo kubectl apply -f k8s/namespace.yaml
sudo kubectl apply -f k8s/config/configmap.yaml
sudo kubectl create secret generic paylater-secret \
  --namespace paylater \
  --from-literal=MYSQL_ROOT_PASSWORD='replace-with-a-strong-password' \
  --from-literal=DB_PASSWORD='replace-with-the-same-mysql-password' \
  --from-literal=JWT_SECRET='replace-with-a-long-random-jwt-secret' \
  --from-literal=INTERNAL_API_TOKEN='replace-with-a-long-random-internal-token' \
  --from-literal=ADMIN_EMAIL='admin@example.com' \
  --from-literal=ADMIN_PASSWORD='replace-with-a-strong-admin-password'
```

Deploy in dependency order:

```bash
sudo kubectl apply -f k8s/database/mysql.yaml
sudo kubectl apply -f k8s/backend/user/user-service.yaml
sudo kubectl apply -f k8s/backend/merchant/merchant-service.yaml
sudo kubectl apply -f k8s/backend/auth/auth-service.yaml
sudo kubectl apply -f k8s/backend/ledger/ledger-service.yaml
sudo kubectl apply -f k8s/backend/report/report-service.yaml
sudo kubectl apply -f k8s/backend/gateway/api-gateway.yaml
sudo kubectl apply -f k8s/frontend/frontend-deployment.yaml
sudo kubectl apply -f k8s/frontend/frontend-service.yaml
sudo kubectl apply -f k8s/ingress/aws-ingress.yaml
```

Verify:

```bash
sudo kubectl get pods -n paylater
sudo kubectl get svc -n paylater
sudo kubectl get ingress -n paylater
```

Open `http://<EC2_PUBLIC_IP>`. The AWS Ingress is hostless and uses Traefik's `StripPrefix` middleware, so `/api/register` becomes `/register` at the gateway. The frontend continues calling `/api`; it never calls internal Kubernetes service names.

Production traffic is same-origin, so the browser does not require CORS for the public frontend-to-gateway requests. `CORS_ALLOWED_ORIGIN` remains configurable in the gateway image for a future cross-origin deployment, and the existing localhost default continues to support local Vite development.

The application images currently use the existing Docker Hub tags (`v1`, `v2`, and `v3`) and are not rebuilt or retagged by this deployment setup. The frontend uses `IfNotPresent`; rebuilds should use a new tag when the image contents change.
