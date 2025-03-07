FROM golang:alpine AS builder
WORKDIR /app
COPY . .

RUN go build -o /telegram-alert-webhook

FROM alpine

WORKDIR /root

COPY --from=builder /telegram-alert-webhook ./app

# Tells Docker which network port your container listens on
EXPOSE 9089

CMD [ "./app" ]
