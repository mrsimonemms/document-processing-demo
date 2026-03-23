FROM golang
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOCACHE=/go/.cache
USER 1000
WORKDIR /go/app
COPY --chown=1000:1000 . .
COPY --from=cosmtrek/air /go/bin/air /go/bin/air
CMD [ "air" ]
