# ===============================
# FRONTEND BUILD STAGE
# ===============================
FROM node:24.6-bookworm AS frontend-builder

# The app sits in /frontend/caregiver so its "../shared" resolves to /frontend/shared.
WORKDIR /frontend/caregiver

ARG VITE_API_URL
ARG VITE_GOOGLE_CLIENT_ID
# Where the caregiver app frames the carepartner web build for its screen preview
# (the carepartner stage below, served by this server at /carepartner/).
ARG VITE_CAREPARTNER_URL
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID
ENV VITE_CAREPARTNER_URL=$VITE_CAREPARTNER_URL

COPY caregiver/package*.json ./
RUN npm install

COPY shared/ /frontend/shared/
COPY caregiver/ ./
RUN npm run build


# ===============================
# CAREPARTNER WEB BUILD STAGE
# ===============================
# The carepartner app's web build, only for the caregiver app's screen preview
# (the tablet app ships separately through EAS). Served at /carepartner/.
FROM node:24.6-bookworm AS carepartner-builder

# The app sits in /frontend/carepartner so its "../shared" resolves to /frontend/shared.
WORKDIR /frontend/carepartner

# The same server address as the caregiver stage: Expo reads it as EXPO_PUBLIC_API_URL.
ARG VITE_API_URL
ENV EXPO_PUBLIC_API_URL=$VITE_API_URL
# Must match webapp.CarepartnerPath in the server.
ENV CAREPARTNER_WEB_BASE_URL=/carepartner

COPY carepartner/package*.json ./
RUN npm ci

COPY shared/ /frontend/shared/
COPY carepartner/ ./
# An empty address would build an app that can't reach the server.
RUN test -n "$EXPO_PUBLIC_API_URL" || (echo "Set the VITE_API_URL build arg" >&2; exit 1)
RUN npx expo export --platform web --output-dir dist


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
COPY --from=frontend-builder /frontend/caregiver/dist /app/frontend
COPY --from=carepartner-builder /frontend/carepartner/dist /app/carepartner

# ffmpeg provides ffprobe, which reads video and audio durations on upload
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates ffmpeg && rm -rf /var/lib/apt/lists/*

EXPOSE 80 443

CMD ["./main"]