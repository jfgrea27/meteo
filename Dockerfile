FROM golang:1.25-alpine AS builder
ARG SVC

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/${SVC} ./cmd/${SVC}

FROM alpine:latest
ARG SVC
ENV SVC=${SVC}

COPY --from=builder /bin/${SVC} /bin/${SVC}
CMD /bin/${SVC}