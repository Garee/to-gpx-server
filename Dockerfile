FROM golang:1.16.5-buster
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN apt-get update && apt-get -y install --no-install-recommends gpsbabel
RUN go build -o main ./src/main.go
EXPOSE 8080
CMD ["./main"]
