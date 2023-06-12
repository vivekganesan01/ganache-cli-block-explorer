FROM golang:alpine3.18 AS builder
RUN mkdir /app
WORKDIR /app
COPY . .
RUN go build -o ganache-cli-block-explorer .

FROM alpine:3.18 AS final
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/ganache-cli-block-explorer .
COPY --from=builder /app/static ./static
COPY --from=builder /app/template ./template
RUN chmod +x ./ganache-cli-block-explorer
EXPOSE 5051
ENTRYPOINT [ "./ganache-cli-block-explorer", "--bind", "0.0.0.0" ]

# docker build -t ganache-cli-block-explorer .ª
# docker run -d --publish 5051:5051 ganache-cli-block-explorer