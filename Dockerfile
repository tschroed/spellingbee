# syntax=docker/dockerfile:1

# lol. no idea dog.
# go build spellingbee.go before building the image.
# docker build -t spellingbee .
# docker run [-d] spellingbee
# if using -d (detach), then use docker logs to inspec the output

# FROM node:18-alpine
FROM golang:1.23
WORKDIR /app
# ADD https://raw.githubusercontent.com/dwyl/english-words/master/words.txt words.txt
# ADD https://www.mit.edu/~ecprice/wordlist.10000 words.txt
ADD https://websites.umich.edu/~jlawler/wordlist words.txt
COPY . .
COPY server/page_html.tmpl .
# Examples of both shell mode and exec mode invocations.
RUN go build -o spellingbee_server server/server.go
CMD ["./spellingbee_server", "words.txt"]
EXPOSE 3000
