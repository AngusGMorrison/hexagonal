FROM golang:1.18

# netcat is a dependency of ./scripts/wait_for.sh
RUN apt-get update && apt-get install -y netcat

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

EXPOSE 3000
