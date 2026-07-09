# Build the frontend. Its output is embedded into the Go binary, so it must
# exist before the Go build runs.
FROM node:22-alpine AS frontend
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig*.json vite.config.ts index.html ./
COPY frontend/ ./frontend/
COPY public/ ./public/
RUN npm run build

FROM golang:1.26 AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY internal/ ./internal/
COPY migrations/ ./migrations/
COPY --from=frontend /app/dist/ ./dist/
# modernc.org/sqlite is pure Go, so the binary is static and needs no libc.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /fuelprices .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=backend /fuelprices /fuelprices

# The SQLite database lives here and must be backed by persistent storage.
VOLUME /data
ENV DB_PATH=/data/fuelprices.db
EXPOSE 3000

ENTRYPOINT ["/fuelprices"]
