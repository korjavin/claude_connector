# Secure Local Data Connector for Claude.ai

## 5.1. Project Overview

This project provides a secure, self-hosted Model Context Protocol (MCP) server that acts as a custom connector for Claude.ai. It is designed to allow Claude to access and process data from a local CSV file on your server without the sensitive data ever leaving your control. The entire application is containerized with Docker for easy, reproducible deployment.

## 5.2. Features

- **Secure Data Access**: Provides read-only access to a local CSV file. The data is processed on your server and only the requested results are sent to Claude.
- **Single Tool**: Exposes one specific tool to Claude: `get_last_n_records`.
- **Authentication**: Uses Ory Hydra, a certified OAuth 2.0 and OpenID Connect provider, for secure, token-based authentication.
- **Self-Hosted**: Designed to run on your own infrastructure using Docker and Docker Compose.

## 5.3. Architecture

The system uses a layered architecture to ensure security and reliability. It includes the Go connector, an Ory Hydra instance for OAuth2, and a consent application. A request from Claude is securely routed through Cloudflare to the Go application, which validates the provided JWT with Hydra before processing the request.

For a detailed breakdown of the components, data flow, and security model, please see [architecture.md](architecture.md).

## 5.4. Configuration

The application is configured using environment variables. Create a `.env` file in the root of the project directory.

### .env file example:

```env
# Port for the MCP server to listen on
MCP_SERVER_PORT=8080

# The path to the CSV file *inside the container*
CSV_FILE_PATH=/data/medical_data.csv
```

The OAuth2 provider (Ory Hydra) is configured via `docker-compose.oauth.yml` and the `scripts/configure-hydra.sh` script.

### Environment Variable Configuration

| Variable Name | Description | Example Value |
|---------------|-------------|---------------|
| MCP_SERVER_PORT | The internal port on which the Go web server will listen. | 8080 |
| CSV_FILE_PATH | The absolute path to the data file as seen from inside the Docker container. | /data/medical_data.csv |

## 5.5. Deployment

### Prerequisites

- Docker and Docker Compose installed on your deployment server.
- A server with a public IP address.
- A custom domain managed by Cloudflare.
- (Optional but recommended) Portainer for managing the Docker deployment.
- GitHub repository with Actions enabled for CI/CD.

### Docker Compose Files

This project includes three docker-compose configurations:

- **`docker-compose.yml`** - Production deployment with GitHub Container Registry images and named volumes
- **`docker-compose.dev.yml`** - Local development with local build and `./data` directory mount
- **`docker-compose.prod.yml`** - Production-ready with health checks and explicit environment variables

### Manual Deployment Steps

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/korjavin/claude_connector.git
   cd claude_connector
   ```

2. **Prepare Data and Configuration**:
   - Place your sensitive CSV file inside the `data/` directory and name it `medical_data.csv`.
   - Create the `.env` file if it doesn't exist.

3. **Deploy with Docker Compose**:
   From the project's root directory, run the following command to bring up the entire stack, including the connector, Ory Hydra, and the consent app:
   ```bash
   docker compose -f docker-compose.oauth.yml -f docker-compose.dev.yml up --build -d
   ```

4. **Configure Cloudflare and Portainer**:
   Follow the detailed instructions in [guide.md](guide.md) to:
   - Set up an A record in Cloudflare pointing your subdomain to your server's IP.
   - (If using Portainer) Deploy the stack using the Portainer UI for easier management.

### Automated CI/CD Deployment

This project includes a complete CI/CD pipeline using GitHub Actions that automatically builds, pushes to GitHub Container Registry (GHCR), and deploys your application.

#### Setup Steps:

1. **Configure Repository Settings**:
   - In your GitHub repository, go to Settings > Actions > General
   - Ensure "Read and write permissions" are enabled for GITHUB_TOKEN

2. **Update docker-compose.yml**:
   Replace the placeholders in `docker-compose.yml`:
   - `korjavin` is already set as the GitHub username
   - `claude-connector.your-domain.com` → Your actual domain
   - `claude-network` → Your Docker network name (optional)

3. **Set up Portainer Webhook (Optional)**:
   - In Portainer, create a webhook for your stack
   - Add the webhook URL as a GitHub secret named `PORTAINER_REDEPLOY_HOOK`

4. **Configure Traefik Labels (Optional)**:
   If using Traefik for reverse proxy, the docker-compose.yml already includes appropriate labels. Adjust as needed for your setup.

#### How it Works:

1. **Trigger**: Push to `main` branch or manual workflow dispatch
2. **Build**: GitHub Actions builds the Docker image with embedded commit SHA
3. **Push**: Image is pushed to GitHub Container Registry (ghcr.io)
4. **Deploy**: Updates `docker-compose.yml` in `deploy` branch with new image tag
5. **Redeploy**: (Optional) Triggers Portainer webhook to pull and restart containers

#### Image Tagging:

- Each build creates two tags:
  - `ghcr.io/korjavin/claude_connector:$COMMIT_SHA` (specific version)
  - `ghcr.io/korjavin/claude_connector:latest` (latest version)

#### Manual Deployment from GHCR:

To deploy a specific version manually:
```bash
docker pull ghcr.io/korjavin/claude_connector:latest
docker compose up -d
```

## 5.6. Usage

1. **Configure an OAuth2 Client in Hydra**:
   - The `docker-compose` setup automatically runs a script to create a client. You can see the client ID and secret in the logs of the `hydra-cli` container: `docker logs hydra-cli`.
   - You will need to provide this Client ID and Secret to Claude in the connector settings.

2. **Add Connector to Claude**:
   - In your Claude.ai settings, navigate to the "Connectors" section.
   - Click "Add custom connector".
   - Enter the full URL for the MCP endpoint: `https://<your-subdomain.your-domain.com>/mcp`.
   - In the advanced settings, provide the OAuth2 Client ID and Secret you obtained from the `hydra-cli` logs.
   - The token endpoint URL will be `https://<your-hydra-domain>/oauth2/token`.

3. **Example Prompt**:
   Once the connector is successfully added, you can use the tool in your conversations with Claude. For example:

   > "Using my medical data connector, please get the last 5 records and summarize them."

## Works Cited

- Connect Claude Code to tools via MCP - Anthropic, accessed September 19, 2025, https://docs.anthropic.com/en/docs/claude-code/mcp
- Model Context Protocol - Wikipedia, accessed September 19, 2025, https://en.wikipedia.org/wiki/Model_Context_Protocol
- Building Custom Connectors via Remote MCP Servers | Anthropic Help Center, accessed September 19, 2025, https://support.anthropic.com/en/articles/11503834-building-custom-connectors-via-remote-mcp-servers
- Claude.ai, accessed September 19, 2025, https://claude.ai/
- Claude MCP Community, accessed September 19, 2025, https://www.claudemcp.com/
- Tool use with Claude - Anthropic API, accessed September 19, 2025, https://docs.anthropic.com/en/docs/build-with-claude/tool-use
- Cloudflare DNS | Authoritative and Secondary DNS, accessed September 19, 2025, https://www.cloudflare.com/application-services/products/dns/
- Custom domains · Cloudflare Pages docs, accessed September 19, 2025, https://developers.cloudflare.com/pages/configuration/custom-domains/
- Build your Go image - Docker Docs, accessed September 19, 2025, https://docs.docker.com/guides/golang/build-images/
- How to Dockerize a Golang Application? - GeeksforGeeks, accessed September 19, 2025, https://www.geeksforgeeks.org/devops/how-to-dockerize-a-golang-application/