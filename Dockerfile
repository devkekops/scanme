FROM golang:alpine3.16 AS build

WORKDIR /app
COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY ./ ./

RUN CGO_ENABLED=0 go build -o /scanmeapp cmd/scanme/main.go

#ARG UID=1000

#RUN adduser \
#    --disabled-password \
#    --no-create-home \
#    --shell /docker-app \
#    --gecos "" \
#    --uid ${UID} \
#    --home / \
#    app

FROM golang:alpine3.16

ENV APP_HOME /go/src/scanmeapp
RUN mkdir -p "$APP_HOME"
WORKDIR "$APP_HOME"

COPY static/ static/
COPY templates/ templates/
COPY --from=build /scanmeapp $APP_HOME

#COPY --from=build /etc/passwd /etc/passwd
#USER app

ENV SERVER_ADDRESS 0.0.0.0:8080
EXPOSE 8080

ENTRYPOINT ["./scanmeapp"]