# Docker Usage for pickemctl

## Building the Image

To build the Docker image locally:

```bash
docker build -t familypickem/pickemctl .
```

## Running the Container

### Daemon Mode (Default)

The container runs in daemon mode by default. You need to provide a `config.yaml` file:

```bash
# Create your config.yaml based on config.yaml.example
cp config.yaml.example config.yaml
# Edit config.yaml with your database settings

# Run the container with mounted config
docker run -d \
  --name pickemctl-daemon \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  familypickem/pickemctl
```

### Running Individual Commands

You can override the default daemon command to run individual operations:

```bash
# Run user statistics
docker run --rm \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  familypickem/pickemctl \
  /app/pickemctl userStats

# Run pick statistics
docker run --rm \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  familypickem/pickemctl \
  /app/pickemctl pickStats

# Run most picked analysis
docker run --rm \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  familypickem/pickemctl \
  /app/pickemctl topPicked

# Run least picked analysis
docker run --rm \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  familypickem/pickemctl \
  /app/pickemctl leastPicked
```

### Using Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  pickemctl:
    image: familypickem/pickemctl:latest
    container_name: pickemctl-daemon
    restart: unless-stopped
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    environment:
      - TZ=America/New_York  # Adjust timezone as needed
```

Run with:
```bash
docker-compose up -d
```

## Configuration

### Required Config.yaml

Your `config.yaml` must include database configuration:

```yaml
# Database configuration
database:
  host: your-postgres-host
  port: 5432
  user: your-username
  password: your-password
  name: pickem
  sslmode: disable

# Application settings
app:
  season:
    current: "2425"

# Daemon settings
daemon:
  interval: 30  # seconds
```

### Using Secrets (Kubernetes/Docker Swarm)

For production deployments, mount the config as a secret:

```bash
# Docker Swarm
docker service create \
  --name pickemctl-daemon \
  --secret source=pickemctl-config,target=/app/config.yaml \
  familypickem/pickemctl

# Kubernetes
kubectl create secret generic pickemctl-config \
  --from-file=config.yaml=./config.yaml

# Then mount in your deployment
```

## Image Details

- **Base Image**: Alpine Linux (minimal footprint)
- **User**: Non-root user (appuser:appgroup, UID/GID 1001)
- **Working Directory**: `/app`
- **Health Check**: Built-in health check using `pickemctl --help`
- **Size**: ~20MB (multi-stage build)

## Logs and Monitoring

View container logs:
```bash
# Follow logs
docker logs -f pickemctl-daemon

# View recent logs
docker logs --tail 100 pickemctl-daemon
```

Check health status:
```bash
docker inspect --format='{{.State.Health.Status}}' pickemctl-daemon
```

## Troubleshooting

### Container won't start
- Verify `config.yaml` is mounted correctly
- Check database connectivity from the container
- Ensure the config file has proper permissions

### Database connection issues
- Verify database host is accessible from container
- Check firewall rules if using external database
- Ensure database credentials are correct

### View container filesystem
```bash
docker exec -it pickemctl-daemon sh
``` 