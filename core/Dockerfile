FROM golang:1.25.5 AS builder

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates ffmpeg

COPY --from=builder server /bin/server
COPY session.json /bin/session.json
COPY .env /bin/.env
COPY configs/values_local.yaml /bin/configs/values_local.yaml

ENV CONFIG_FILE=/bin/configs/values_local.yaml
ENV PROJECT_DEBUG=true
ENV PROJECT_AUDIO_PATH=/bin/audio

WORKDIR /bin

ENTRYPOINT ["/bin/server"]