# Build React frontend using latest Node.js
FROM node:18 AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Build Go backend using latest Go version
FROM golang:1.20 AS backend-builder
WORKDIR /backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go build -o app .

# Final image to serve both frontend and backend
FROM golang:1.20

# Set working directory
WORKDIR /app

# Copy Go app and React build
COPY --from=frontend-builder /frontend/build /app/static
COPY --from=backend-builder /backend/app /app/

# Expose port
EXPOSE 8080
CMD ["./app"]
