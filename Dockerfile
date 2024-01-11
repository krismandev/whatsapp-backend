#builder
FROM golang:alpine AS builder
ARG VER
ARG BUILDDATE
RUN apk update && apk add --no-cache git
RUN apk --no-cache add ca-certificates
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
RUN go get -d -v
RUN go build -ldflags="-w -s -X 'main.version=$VER' -X 'main.builddate=$BUILDDATE'" -o /go/bin/skeleton

#Real Image
FROM alpine
RUN apk add --no-cache tzdata
RUN mkdir /app && mkdir /app/config && mkdir -p /app/data/log
RUN addgroup -g 1010 -S app
RUN adduser -S --ingroup app -u 1010 app 
RUN chown -R app:app /app
ENV TZ Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
USER app
WORKDIR /app
COPY --chown=app --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --chown=app --from=builder /go/bin/skeleton /app/skeleton.app
COPY --chown=app config/config.yml /app/config/config.yml
CMD /app/skeleton.app
