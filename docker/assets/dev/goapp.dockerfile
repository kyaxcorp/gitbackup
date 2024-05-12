# syntax=docker/dockerfile:1

FROM golang:1.22

RUN go install github.com/cosmtrek/air@latest
RUN go install github.com/maruel/panicparse/v2@latest
