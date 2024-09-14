# syntax=docker/dockerfile:1

# lol. no idea dog.
# go build spellingbee.go before building the image.
# docker build -t spellingbee .
# docker run [-d] spellingbee
# if using -d (detach), then use docker logs to inspec the output

FROM node:18-alpine
WORKDIR /app
ADD https://raw.githubusercontent.com/dwyl/english-words/master/words.txt words.txt
COPY . .
RUN yarn install --production
CMD ["time", "./spellingbee", "words.txt", "ajvelin"]
EXPOSE 3000
