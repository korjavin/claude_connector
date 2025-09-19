# Deployment Guide

This guide provides step-by-step instructions for deploying the Claude Connector using the automated CI/CD pipeline.

## Prerequisites

1. **GitHub Repository**: Your code should be in a GitHub repository
2. **Docker Server**: A server with Docker and Docker Compose installed
3. **Domain**: A domain name managed by Cloudflare (optional but recommended)
4. **Portainer**: Recommended for easier Docker management

## Quick Start

### 1. Fork/Clone Repository

```bash
git clone https://github.com/korjavin/claude_connector.git
cd claude_connector
```

### 2. Configure Repository

Replace placeholders in `docker-compose.yml`:

```yaml
# Already configured for korjavin:
image: ghcr.io/korjavin/claude_connector:latest
```

Update domain in Traefik labels:
```yaml
# Change this:
- "traefik.http.routers.claude-connector.rule=Host(`claude-connector.your-domain.com`)"

# To your domain:
- "traefik.http.routers.claude-connector.rule=Host(`claude-connector.yourdomain.com`)"
```

### 3. Set Up GitHub Secrets (Optional)

In your GitHub repository settings:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add repository secret:
   - `PORTAINER_REDEPLOY_HOOK`: Your Portainer webhook URL (if using Portainer)

### 4. Enable GitHub Actions

1. Go to **Settings** → **Actions** → **General**
2. Set **Workflow permissions** to "Read and write permissions"
3. Save

### 5. Deploy

Push to main branch:
```bash
git add .
git commit -m "Configure deployment"
git push origin main
```

The GitHub Actions workflow will:
1. Build the Docker image
2. Push it to GitHub Container Registry (GHCR)
3. Update the deploy branch with new image tag
4. Trigger Portainer redeploy (if configured)

## Server Setup

### Option 1: Using Portainer (Recommended)

1. **Install Portainer** on your server
2. **Create Stack** from the `docker-compose.yml`
3. **Set Environment Variables in Portainer**:
   In the Stack deployment screen, scroll down to "Environment variables" and add:
   - `API_SECRET_KEY`: Your secure random string (REQUIRED)
   - `CLAUDE_DOMAIN`: Your domain (e.g., claude-connector.yourdomain.com)
   - `TLS_RESOLVER`: Your Traefik TLS resolver name (e.g., myresolver)
   - `NETWORK_NAME`: Your Docker network name (e.g., claude-network)
   - `MCP_SERVER_PORT`: 8080 (optional, has default)
   - `CSV_FILE_PATH`: /data/medical_data.csv (optional, has default)

4. **Create Data Directory**:
   ```bash
   sudo mkdir -p /opt/claude-connector/data
   sudo chown -R $USER:$USER /opt/claude-connector/data
   # Copy your CSV file to this location
   cp your-medical-data.csv /opt/claude-connector/data/medical_data.csv
   ```

5. **Volume Configuration**:
   The docker-compose.yml uses a named volume that binds to `/opt/claude-connector/data` on the host

### Option 2: Direct Docker Compose

1. **Clone on Server**:
   ```bash
   git clone https://github.com/korjavin/claude_connector.git
   cd claude_connector
   git checkout deploy  # Use the deploy branch for latest image
   ```

2. **Create .env file**:
   ```bash
   # Copy the production template
   cp .env.prod.example .env

   # Edit the configuration
   cat > .env << EOF
   MCP_SERVER_PORT=8080
   API_SECRET_KEY=your-very-long-secure-random-string-here
   CSV_FILE_PATH=/data/medical_data.csv
   CLAUDE_DOMAIN=claude-connector.yourdomain.com
   TLS_RESOLVER=myresolver
   EOF
   ```

3. **Deploy**:
   ```bash
   docker compose pull
   docker compose up -d
   ```

## Traefik Configuration (Optional)

If using Traefik for reverse proxy and SSL:

1. **Create Network**:
   ```bash
   docker network create claude-network
   ```

2. **Traefik Labels** are already configured in `docker-compose.yml`

3. **Update Domain**: Replace `claude-connector.your-domain.com` with your domain

## Monitoring

### Health Check

The application includes a health endpoint:
```bash
curl http://your-server:8080/health
```

Response:
```json
{
  "status": "ok",
  "commit": "abc123def456",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Logs

View logs with:
```bash
docker compose logs -f claude-connector
```

## CI/CD Workflow

### Automatic Deployment

1. **Trigger**: Push to `main` branch
2. **Build**: Creates Docker image with commit SHA
3. **Push**: Uploads to `ghcr.io/username/claude_connector`
4. **Update**: Updates `deploy` branch with new image tag
5. **Deploy**: Triggers Portainer webhook (optional)

### Manual Deployment

Trigger manually in GitHub:
1. Go to **Actions** tab
2. Select **Deploy Claude Connector** workflow
3. Click **Run workflow**

### Rollback

To rollback to a previous version:
1. Find the commit SHA of the previous version
2. Update docker-compose.yml manually:
   ```yaml
   image: ghcr.io/username/claude_connector:PREVIOUS_COMMIT_SHA
   ```
3. Redeploy

## Security Considerations

1. **Environment Variables**: Never commit the `.env` file
2. **API Keys**: Use strong, random API keys
3. **Network**: Use Docker networks for isolation
4. **Firewall**: Only expose necessary ports
5. **Updates**: Regularly update base images and dependencies

## Troubleshooting

### Build Failures

Check GitHub Actions logs:
1. Go to **Actions** tab
2. Click on failed workflow
3. Review build logs

### Deployment Issues

1. **Check image exists**:
   ```bash
   docker pull ghcr.io/korjavin/claude_connector:latest
   ```

2. **Verify environment variables**:
   ```bash
   docker compose config
   ```

3. **Check logs**:
   ```bash
   docker compose logs claude-connector
   ```

### Common Issues

1. **Permission denied**: Check GitHub token permissions
2. **Image not found**: Verify repository name in docker-compose.yml
3. **Health check failed**: Check if port 8080 is accessible
4. **CSV file not found**: Verify volume mount and file path
5. **"FATAL: API_SECRET_KEY environment variable not set"**:
   - In Portainer: Ensure `API_SECRET_KEY` is set in the Environment variables section
   - For docker-compose: Ensure the .env file exists and contains `API_SECRET_KEY=your-key`
   - Check container logs: `docker logs claude-connector-service`

## Production Checklist

- [ ] Repository configured with correct image name
- [ ] GitHub Actions permissions enabled
- [ ] API_SECRET_KEY set to secure random value
- [ ] Medical data CSV file uploaded to server
- [ ] Domain DNS configured (if using Traefik)
- [ ] SSL certificates configured
- [ ] Firewall rules configured
- [ ] Monitoring and logs set up
- [ ] Backup strategy for data directory