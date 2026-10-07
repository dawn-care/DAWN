# ===============================
# FRONTEND BUILD STAGE
# ===============================
FROM node:24.6-bookworm AS frontend-builder

WORKDIR /frontend

ARG VITE_API_URL
ARG VITE_GOOGLE_CLIENT_ID
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID

COPY caregiver/package*.json ./
RUN npm install

COPY caregiver/ ./
RUN npm run build


# ===============================
# BACKEND BUILD STAGE
# ===============================
FROM golang:1.25.3-bookworm AS builder

WORKDIR /app

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g ./cmd/main.go -o ./cmd/docs

RUN go build -o main ./cmd && go build -o migrate ./cmd/migrate


# ===============================
# RUNTIME IMAGE
# ===============================
FROM ubuntu:24.04

WORKDIR /app

# Copy Go binary
COPY --from=builder /app/main /app/migrate ./

# Copy built frontend
COPY --from=frontend-builder /frontend/dist /app/frontend

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

EXPOSE 80 443

CMD ["./main"]