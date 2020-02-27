FROM golang AS builder
WORKDIR /go/src/github.com/eniot/sensor-data-collector
COPY . .
RUN go get -d .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

FROM scratch
WORKDIR /app
EXPOSE 80
COPY --from=builder /go/src/github.com/eniot/sensor-data-collector/sensor-data-collector .
ENTRYPOINT [ "/app/sensor-data-collector"]