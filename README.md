# PostgreSQL Backup to S3 with Docker

This application automates the process of backing up PostgreSQL databases and uploading them to an S3-compatible storage service, utilizing Docker for easy deployment and scheduling.

## Features

- Easy deployment with Docker and Docker Compose.
- Support for multiple PostgreSQL databases.
- Customizable backup intervals.
- Direct upload of backups to an S3-compatible storage bucket.
- Environment variable and command-line configuration for flexibility.
- Secure handling of database and S3 credentials.

## Prerequisites

- Docker and Docker Compose installed on your system.
- Access to a PostgreSQL database.
- Access to an S3-compatible storage service.

## Configuration

Before running the application, you need to configure it either by setting environment variables in a `.env` file or by using command-line flags in the `docker-compose.yml`.

### Environment Variables

Create a `.env` file in the project directory with the following variables:

- `URLS`: Comma-separated list of PostgreSQL database URLs to backup. Format: `postgres://<user>:<password>@<host>[:<port>]/<dbname>`
- `S3_ENDPOINT`: The endpoint URL of your S3-compatible storage service.
- `S3_BUCKET`: The name of the bucket where backups will be stored.
- `S3_ACCESS_KEY`: Your S3 access key.
- `S3_SECRET_KEY`: Your S3 secret key.
- `INTERVAL`: How often to run the backup (e.g., `24h` for daily backups).

### Docker Compose

Alternatively, you can specify the configuration directly in the `docker-compose.yml` file under the `environment` section of your service:

```yaml
services:
  app:
    build: .
    environment:
      URLS: "postgres://user:password@host:port/dbname"
      S3_ENDPOINT: "your_s3_endpoint"
      S3_BUCKET: "your_s3_bucket"
      S3_ACCESS_KEY: "your_s3_access_key"
      S3_SECRET_KEY: "your_s3_secret_key"
      INTERVAL: "24h"
    volumes:
      - .:/app
```

## Running the Application with Docker

1. Build the Docker image:

   ```sh
   docker compose build
   ```

2. Start the application:

   ```sh
   docker compose up -d
   ```

This will start the application in the background. It will automatically perform backups based on the configured interval and upload them to the specified S3 bucket.

## Monitoring and Logs

To monitor the application's activity and view logs:

```sh
docker compose logs -f
```

This command will follow the log output of the container. Press `Ctrl+C` to exit log following.

## Updating Configuration

If you need to update the configuration, modify the `.env` file or the `docker-compose.yml` as necessary and restart the service:

```sh
docker compose down
docker compose up -d
```
