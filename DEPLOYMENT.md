# Deployment Guide

This guide provides step-by-step instructions for deploying the Claude Connector and its associated authentication services using the provided Docker Compose files.

## Prerequisites

1. **Docker and Docker Compose**: Ensure your server has the latest versions installed.
2. **Public Server**: A server with a public IP address.
3. **Domain Name**: A custom domain is required for exposing the services securely via HTTPS.

## Deployment Options

This project includes two main deployment configurations:

- **Local/Development**: Uses `docker-compose.dev.yml` and `docker-compose.oauth.yml`. Ideal for testing and development. It mounts local source code for live-reloading.
- **Production**: Uses `docker-compose.prod.yml` and `docker-compose.oauth.yml`. This setup is optimized for production use, using pre-built images from a container registry.

## Development Deployment

### 1. Clone the Repository

```bash
git clone https://github.com/korjavin/claude_connector.git
cd claude_connector
```

### 2. Prepare Data and Environment

- Place your sensitive CSV file in the `./data` directory (e.g., `./data/medical_data.csv`).
- Create a `.env` file if one does not exist. The only required variable for the connector itself is `CSV_FILE_PATH`, which defaults to `/data/medical_data.csv`.

### 3. Launch the Stack

This command will start all services, including the `claude-connector`, `hydra`, `hydra-db`, and the `consent` app.

```bash
docker compose -f docker-compose.oauth.yml -f docker-compose.dev.yml up --build -d
```

### 4. Configure the OAuth2 Client

After a few moments, the `hydra-cli` service will run a script to create an OAuth2 client. Retrieve the client ID and secret with the following command:

```bash
docker logs hydra-cli
```

Save the `client_id` and `client_secret` from the output.


## Production Deployment (CI/CD)

The production deployment is designed to be automated via the CI/CD pipeline in `.github/workflows/deploy.yml`.

#### Portainer Webhook Setup (Recommended)

**Step 1: Create Webhook in Portainer**
1. In Portainer, navigate to your deployed stack
2. Click on the stack name to view details
3. Go to the "Webhooks" tab
4. Click "Add webhook"
5. Give it a name (e.g., "github-auto-deploy")
6. Copy the generated webhook URL (format: `https://your-portainer.domain/api/webhooks/xxx-xxx-xxx`)

**Step 2: Add Secret to GitHub Repository**
1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Name: `PORTAINER_REDEPLOY_HOOK`
4. Value: The webhook URL from Portainer
5. Click **Add secret**

With this setup, Portainer will automatically pull and redeploy your containers when GitHub Actions builds new images.


### 1. Fork the Repository

Fork the `korjavin/claude_connector` repository to your own GitHub account.

### 2. Configure Your Domain

In `docker-compose.prod.yml`, update the Traefik labels to use your domain:

```yaml
services:
  claude-connector:
    # ...
    labels:
      - "traefik.http.routers.claude-connector.rule=Host(`claude-connector.your-domain.com`)"
      # ...
  hydra:
    # ...
    labels:
      - "traefik.http.routers.hydra.rule=Host(`hydra.your-domain.com`)"
      # ...
```

### 3. Set Up GitHub Secrets

In your forked repository's settings (**Settings** > **Secrets and variables** > **Actions**), add the following secrets:

- `DOCKERHUB_USERNAME`: Your Docker Hub username.
- `DOCKERHUB_TOKEN`: Your Docker Hub access token.
- `SERVER_HOST`: The IP address or hostname of your deployment server.
- `SERVER_USER`: The SSH username for your server.
- `SSH_PRIVATE_KEY`: The SSH private key to access your server.

### 4. Trigger the CI/CD Pipeline

Push a commit to the `main` branch. The GitHub Actions workflow will:
1.  Build and push the `claude-connector` Docker image to Docker Hub.
2.  Connect to your server via SSH.
3.  Run `docker-compose` using the `docker-compose.prod.yml` and `docker-compose.oauth.yml` files to deploy the entire stack.

### 5. Configure the OAuth2 Client in Production

SSH into your server and run `docker logs hydra-cli` to get the production client ID and secret.

## Connecting to Claude.ai

1.  In your Claude.ai settings, add a new custom connector.
2.  Set the **URL** to your public MCP endpoint (e.g., `https://claude-connector.your-domain.com/mcp`).
3.  Set the **Authentication** to "OAuth 2.0" with the "Client Credentials" grant type.
4.  Provide the **Client ID** and **Client Secret** you obtained from the `hydra-cli` logs.
5.  Set the **Token URL** to your public Hydra token endpoint (e.g., `https://hydra.your-domain.com/oauth2/token`).
6.  Save and enable the connector.