# syntax=docker/dockerfile:1

# lol. no idea dog.
# go build spellingbee.go before building the image.
# docker build -t spellingbee .
# docker run [-d] spellingbee
# if using -d (detach), then use docker logs to inspec the output

# FROM node:18-alpine
FROM golang:1.23
WORKDIR /app
ADD https://raw.githubusercontent.com/dwyl/english-words/master/words.txt words.txt
COPY . .
# Examples of both shell mode and exec mode invocations.
RUN go build spellingbee.go
CMD ["./spellingbee", "words.txt", "ajvelin"]
EXPOSE 3000
