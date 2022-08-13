FROM golang:alpine as build
WORKDIR /app
COPY . .
RUN go build

FROM alpine:latest
COPY --from=build /app/jaf /app/jaf
WORKDIR /app
CMD ["./jaf"]
