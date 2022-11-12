FROM golang:1.19-alpine AS builder

ADD . /app
WORKDIR /app
RUN go build .

FROM alpine as runner

WORKDIR /app
ADD templates /app/templates
COPY --from=builder /app/edulink /app/edulink

CMD ["/app/edulink"]