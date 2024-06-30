# Build stage
FROM golang:1.22.3-alpine3.19 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
# RUN apk add curl 
# RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz # เนื่องจากไม่จำเป็นต้อง download แล้ว

# Run stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main .
# COPY --from=builder /app/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./db/migration

EXPOSE 8080 
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]