FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

EXPOSE ${orchestrator_port}
EXPOSE ${agent_port}
EXPOSE 8080

COPY ./src ./src
COPY ./starter ./starter