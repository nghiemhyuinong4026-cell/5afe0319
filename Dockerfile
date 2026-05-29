FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o main .

FROM postgres:15-bookworm

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main .
COPY .env /app/.env

RUN mkdir -p /var/lib/postgresql/15/main && chown -R postgres:postgres /var/lib/postgresql

RUN git init && git add -A && git config user.email "docker@local" && git config user.name "Docker" && git commit -m "Initial commit" --allow-empty

EXPOSE 8080

CMD if [ ! -s /var/lib/postgresql/15/main/PG_VERSION ]; then \
    su - postgres -c "/usr/lib/postgresql/15/bin/initdb -D /var/lib/postgresql/15/main"; fi && \
    su - postgres -c "/usr/lib/postgresql/15/bin/pg_ctl -D /var/lib/postgresql/15/main -o '-c listen_addresses=localhost' start" && \
    sleep 2 && \
    su - postgres -c "/usr/lib/postgresql/15/bin/createdb vehicle_management" || true && \
    ./main