# End-to-End Implementation and Deployment Guide

## 2.1. Overview

This guide provides a complete walkthrough for setting up, deploying, and using the Secure Local Data Connector for Claude.ai. The system uses a multi-container Docker setup, managed by Docker Compose, that includes the main Go connector application and a secure OAuth 2.0 provider (Ory Hydra).

## 2.2. Prerequisites

- **Docker & Docker Compose**: The Docker Engine and Docker Compose tool are required for containerization and orchestration.
- **Server Access**: A server with a public IP address where the application will be deployed.
- **Custom Domain**: A domain name managed by a service like Cloudflare is recommended for secure HTTPS access.

## 2.3. System Components

The deployed system consists of four main services orchestrated by Docker Compose:

1.  **`claude-connector`**: The core Go application that receives requests from Claude, authenticates them, and interacts with the local data file.
2.  **`hydra`**: The Ory Hydra service, which acts as the OAuth 2.0 and OpenID Connect provider. It is responsible for issuing access tokens.
3.  **`hydra-db`**: A PostgreSQL database used by Hydra to persist its configuration and state.
4.  **`consent`**: A simple, pre-built login and consent UI that Hydra uses to authenticate users and grant access to clients.

## 2.4. Deployment Steps

### Step 1: Clone the Repository

```bash
git clone https://github.com/korjavin/claude_connector.git
cd claude_connector
```

### Step 2: Prepare Your Data

Place your sensitive CSV file inside the `data/` directory. The application is configured to look for a file named `medical_data.csv` by default.

```bash
# Example:
cp /path/to/your/data.csv ./data/medical_data.csv
```

### Step 3: Launch the Services

This project uses a multi-file Docker Compose setup. `docker-compose.oauth.yml` defines the authentication services (Hydra, etc.), and `docker-compose.dev.yml` defines the connector service.

From the project's root directory, run the following command to build and start all containers:

```bash
docker compose -f docker-compose.oauth.yml -f docker-compose.dev.yml up --build -d
```

This command will:
- Pull the required Docker images for PostgreSQL, Ory Hydra, and the consent app.
- Build the `claude-connector` image from the local Dockerfile.
- Create a shared Docker network (`hydra-network`) for the services to communicate.
- Start all services in detached mode (`-d`).

### Step 4: Configure the OAuth2 Client

Ory Hydra needs to be configured with an OAuth2 client that Claude.ai can use to authenticate. A script is provided to automate this. The `docker-compose.oauth.yml` file starts a temporary `hydra-cli` container that runs this script.

- **Wait a few moments** for the services to initialize. The script will wait for Hydra to be ready before running.
- **Retrieve Client Credentials**: The script will create a new OAuth2 client and print its details to the logs. To get the `client-id` and `client-secret`, run:

  ```bash
  docker logs hydra-cli
  ```

- You will see JSON output containing the client details. **Copy the `client_id` and `client_secret` values**—you will need them to configure the connector in Claude.ai.

  *Example Output:*
  ```json
  {
    "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "client_secret": "your-client-secret-is-here",
    ...
  }
  ```

## 2.5. Connecting to Claude.ai

1.  **Navigate to Claude.ai**: Go to your Claude.ai settings and find the "Connectors" section.
2.  **Add Custom Connector**: Click "Add custom connector".
3.  **Configure the Endpoint**:
    - **URL**: Enter the full, public URL for your connector's MCP endpoint (e.g., `https://claude-connector.yourdomain.com/mcp`).
    - **Authentication**: Select "OAuth 2.0" with the "Client Credentials" grant type.
4.  **Provide Credentials**:
    - **Client ID**: Paste the `client_id` you copied from the `hydra-cli` logs.
    - **Client Secret**: Paste the `client_secret` you copied.
    - **Token URL**: Enter the public URL for Hydra's token endpoint (e.g., `https://your-hydra-domain.com/oauth2/token`).
5.  **Save and Enable**: Save the connector. Claude will now be able to use it.

## 2.6. System Troubleshooting

### "Connector failed to connect" in Claude

- **Check Container Logs**: Use `docker compose logs <service-name>` to inspect the logs of each service.
  - `docker compose logs claude-connector`: Check for errors in the Go application.
  - `docker compose logs hydra`: Check for errors in the Hydra service.
- **Verify Network**: Ensure all services are running on the same Docker network (`hydra-network`) by inspecting them with `docker network inspect hydra-network`.
- **Test Hydra Endpoint**: From your server, try to `curl http://localhost:4444/health/ready`. It should return `{"status":"ok"}`.

### 401 Unauthorized Error

- This means the access token provided by Claude was invalid.
- **Check Hydra Logs**: Look for token validation errors in the `hydra` service logs.
- **Verify JWKS URL**: The Go application is hardcoded to use `http://hydra:4444/.well-known/jwks.json` to get the keys for validating tokens. Ensure the `hydra` service is accessible at that address from within the Docker network.
- **Regenerate Client**: You can stop the stack (`docker compose down`), remove the `hydra-db-data` volume (`docker volume rm <volume_name>`), and start it again to get a fresh client.