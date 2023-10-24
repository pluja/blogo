# Builder stage
FROM devopsworks/golang-upx:latest as builder
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
COPY ./tailwind.config.js .
COPY ./package.json .
COPY ./package-lock.json .
COPY ./static ./static
COPY ./templates ./templates
RUN npm install && \
    npx tailwindcss -i ./static/css/input.css -o ./static/css/style.css --minify

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

COPY --from=node /app/static/css/style.css ./static/css/style.css
ENV PATH="/app:$PATH"

EXPOSE 3000
CMD ["/app/blogo", "-path", "/app"]
