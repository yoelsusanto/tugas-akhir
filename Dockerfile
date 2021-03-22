# Builder
FROM golang:1.15.8-alpine3.13 as build

ENV TARGET_OS linux
ENV TARGET_ARCH amd64

RUN apk update && apk add --no-cache ca-certificates

RUN mkdir -p /app

WORKDIR /app

COPY . .

RUN go mod download
RUN go mod verify
RUN GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} go build -o main

# -------------------------------------------------------------------------
# Runner
FROM golang:1.15.8-alpine3.13

ENV APP_USER user
ENV APP_HOME /app

# Add Group and User
RUN addgroup "${APP_USER}"
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --ingroup "${APP_USER}" \
    --no-create-home \
    "${APP_USER}"

# Setup workdir
RUN mkdir -p /usr/src/app
WORKDIR ${APP_HOME}

# Import certs
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import app
COPY --from=build /app/main ${APP_HOME}/main

RUN apk add libcap && setcap 'cap_net_bind_service=+ep' ${APP_HOME}/main

USER ${APP_USER}
CMD ["./main"]