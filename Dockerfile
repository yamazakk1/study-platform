FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /study-platform ./cmd/api

FROM alpine:3.22
WORKDIR /app
COPY --from=build /study-platform /app/study-platform
COPY web /app/web
EXPOSE 8080
CMD ["/app/study-platform"]
