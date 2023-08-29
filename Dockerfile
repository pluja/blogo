# Builder stage
FROM devopsworks/golang-upx:1.20 as builder
ENV DEBIAN_FRONTEND noninteractive
WORKDIR /app
COPY blogo .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o blogo . && \
    upx blogo

RUN chmod a+rx blogo

RUN mkdir -p /app/articles/

# Node stage
FROM node:alpine AS node
WORKDIR /app
COPY . .
RUN npm install && \
    npx tailwindcss -o ./style.css --minify

# Start from a complete image
FROM alpine:latest as certs
RUN apk --update add ca-certificates

# Final stage
FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=builder /app/blogo /app/blogo
COPY --from=builder /app/articles /app/articles

# Copy HTML files
COPY templates templates
COPY static static

COPY --from=node /app/style.css ./html/static/css/style.css

EXPOSE 3000
CMD ["/app/blogo", "-path", "/app"]
