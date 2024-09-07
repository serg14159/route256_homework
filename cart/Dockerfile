FROM golang:1.22.5-alpine as builder

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /cart ./cmd/main/main.go

FROM scratch
COPY --from=builder cart /bin/cart

ENTRYPOINT ["/bin/cart"]