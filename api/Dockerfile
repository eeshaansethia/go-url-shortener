FROM golang:alpine AS builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go mod -o main.

FROM alpine
RUN adduser -S -D -H /app appuser
USER appuser
COPY . /app
COPY --from=builder /build/main /app/
WORKDIR /app
EXPOSE 3000
CMD ["./main"]