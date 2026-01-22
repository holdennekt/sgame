# SGAME — Hot to run

This guide provides instructions on how to set up and run the project locally using two different methods: **Docker Compose** (for quick development) or **Kubernetes (Minikube)** (for production-like testing).

## Prerequisites

Before you begin, ensure you have 1 of the following installed:

* **Docker Desktop** (for Docker Compose)
* **Minikube** + **kubectl** (for Kubernetes).

## Docker Compose:

The easiest way to get the environment up and running with a single command.

1. Create a `.env` file from the provided template and fill in the values:

```bash
cp .env.template .env
```

2. Execute the following command in the project root directory:

```
docker-compose up --build
```

Testing

* **Frontend:** **http://localhost:3000**
* **Backend API:** **http://localhost:8080/api**
* **Logs:** `docker-compose logs -f`

## Kubernetes

Use this method to test Ingress, load balancing, and WebSockets in a cluster environment.

1. Install the required tools using Homebrew (macOS) or your preferred package manager:

```
brew install minikube kubectl hyperkit
```

2. Initialize the cluster using the `hyperkit` driver for better networking:

```
minikube start --driver=hyperkit --disk-size=5g --memory=4096
minikube addons enable ingress
```

3. Build images

```
minikube image build -t sgame-backend:local backend
minikube image -t sgame-frontend:local frontend
```

4. Create a secrets file from the provided template and fill in the values:

```
cp k8s/secret.template.yaml k8s/secret.yaml 
```

5. Apply the Kubernetes manifests:

```
kubectl apply -f secret.yaml -f mongo -f redis -f backend -f frontend -f ingress.yaml
```

6. Get the Minikube IP address:

```
minikube ip
```

7. Map the domain in your `/etc/hosts` file:

```
<MINIKUBE-IP> sgame.local
```

Testing

* **Frontend:** **http://sgame.local**
* **Backend API:** **http://sgame.local/api**
* **Deployment Logs:** `kubectl logs -f -l app=sgame-backend --prefix`
