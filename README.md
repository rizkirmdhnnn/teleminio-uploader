# Teleminio Uploader

Teleminio Uploader is a Telegram bot that automatically downloads media files from specified users and uploads them to MinIO storage. It organizes the files by username and media type, making it easy to manage and access media content.

## Features

- Automatic media download from Telegram messages
- Direct upload to MinIO storage
- Organized file structure by username and media type
- Docker support for easy deployment
- Configurable user targeting
- Session persistence

## Prerequisites

- Docker and Docker Compose
- MinIO server instance
- Telegram API credentials (App ID and Hash)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/rizkirmdhnnn/teleminio-uploader.git
   cd teleminio-uploader
   ```

2. Copy the example environment file and configure your settings:
   ```bash
   cp .env.example .env
   ```

3. Edit the `.env` file with your credentials:
   ```env
   APP_ID=your_telegram_app_id
   APP_HASH=your_telegram_app_hash
   PHONE=your_phone_number
   USER_TARGET=username1,username2
   MINIO_ENDPOINT=your_minio_endpoint
   MINIO_ACCESS_KEY=your_minio_access_key
   MINIO_SECRET_KEY=your_minio_secret_key
   MINIO_BUCKET=your_bucket_name
   MINIO_USE_SSL=true_or_false
   ```

## Running with Docker

1. For first-time setup (requires authentication):
   ```bash
   docker-compose run --rm -it bot
   ```
   This command runs the bot in interactive mode, allowing you to authenticate with Telegram.

2. For subsequent runs:
   ```bash
   docker-compose up -d
   ```

3. View logs:
   ```bash
   docker-compose logs -f
   ```

4. Stop the service:
   ```bash
   docker-compose down
   ```

## Configuration

### Environment Variables

- `APP_ID`: Your Telegram application ID
- `APP_HASH`: Your Telegram application hash
- `PHONE`: Phone number for Telegram authentication
- `USER_TARGET`: Comma-separated list of Telegram usernames to monitor
- `MINIO_ENDPOINT`: MinIO server endpoint
- `MINIO_ACCESS_KEY`: MinIO access key
- `MINIO_SECRET_KEY`: MinIO secret key
- `MINIO_BUCKET`: MinIO bucket name
- `MINIO_USE_SSL`: Whether to use SSL for MinIO connection

### Storage Structure

Files are stored in MinIO with the following structure:
```
{bucket_name}/
  ├── {username}/
  │   ├── photo/
  │   │   └── {filename}
  │   ├── video/
  │   │   └── {filename}
  │   └── document/
  │       └── {filename}
```

## Development

### Requirements

- Go 1.21 or higher
- MinIO server
- Telegram API credentials

### Local Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run the application:
   ```bash
   go run cmd/bot/main.go
   ```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
