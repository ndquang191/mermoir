# Memoir

A couple's photo diary — create dated entries, write your story, and attach photos that get automatically processed into thumbnails in the background.

## Stack

- **Frontend**: React 18 + Vite + TanStack Router + TanStack Query (TypeScript)
- **API**: Go + Gin
- **Worker**: Go (goroutine pool consuming from RabbitMQ)
- **Queue**: RabbitMQ (image processing jobs with dead-letter queue)
- **Database**: PostgreSQL 16
- **Containers**: Docker + Docker Compose

## Project Structure

```
mermoir/
├── api/          # Go API server (Gin)
│   ├── db/       # PostgreSQL connection & schema
│   ├── handlers/ # HTTP handlers (entries, uploads)
│   ├── models/   # Data types
│   └── queue/    # RabbitMQ publisher
├── worker/       # Go image processing worker
├── ui/           # React frontend
│   └── src/
│       ├── api/        # Axios client + types
│       ├── components/ # EntryCard, PhotoGrid
│       └── routes/     # TanStack Router file-based routes
└── docker-compose.yml
```

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Run everything

```bash
docker-compose up --build
```

This starts:
| Service   | URL                          |
|-----------|------------------------------|
| UI        | http://localhost:3000        |
| API       | http://localhost:8080        |
| RabbitMQ  | http://localhost:15672 (guest UI, login: memoir/memoir) |
| PostgreSQL| localhost:5432               |

### Development (local, without Docker)

Copy the example env file and adjust as needed:

```bash
cp .env.example .env
```

Run the API:

```bash
cd api
go mod tidy
go run .
```

Run the worker:

```bash
cd worker
go mod tidy
go run .
```

Run the UI:

```bash
cd ui
npm install
npm run dev
```

## How It Works

1. User creates an entry (date + story) via the UI.
2. User attaches photos — each is saved to `/storage/raw/` and a job is published to RabbitMQ.
3. The worker pool picks up jobs, generates a 300px thumbnail and a compressed full image, then updates the photo status to `ready` in PostgreSQL.
4. The UI polls (or refetches) entries — photos show a spinner while `pending` and display the thumbnail once `ready`.
