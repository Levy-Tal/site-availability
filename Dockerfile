# Build React frontend
FROM node:18 AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install --omit=dev
COPY frontend/ ./
RUN npm run build

# Build Go backend
FROM golang:1.23 AS backend-builder
WORKDIR /backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o site-availability .

# Final minimal image using scratch
FROM scratch
WORKDIR /app
COPY --from=backend-builder /backend/site-availability /app/site-availability
COPY --from=frontend-builder /frontend/build /app/static
USER 65532
EXPOSE 8080
ENTRYPOINT ["/app/site-availability"]
