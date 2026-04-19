FROM node:22-alpine AS ui-builder
WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/ui/dist ./ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-builder /server /server
COPY --from=go-builder /app/migrations /migrations
EXPOSE 8080
CMD ["/server"]
