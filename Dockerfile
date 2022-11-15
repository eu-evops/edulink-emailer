FROM golang:1.19-alpine AS builder

ADD . /app
WORKDIR /app
RUN go build .

FROM alpine as runner

WORKDIR /app
ADD templates /app/templates
ADD site /app/site
ADD public /app/public
COPY --from=builder /app/edulink /app/edulink

EXPOSE 8080

CMD ["/app/edulink"]