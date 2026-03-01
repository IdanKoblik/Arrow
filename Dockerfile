FROM node:22-alpine AS frontend
WORKDIR /build/ui
COPY ui/package*.json ./
RUN npm ci --silent
COPY ui/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o arrow ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app
COPY --from=backend /build/arrow ./arrow
COPY --from=frontend /build/ui/dist ./ui/dist

EXPOSE 8080
CMD ["./arrow"]
