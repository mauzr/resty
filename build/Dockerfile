FROM golang:latest AS build
COPY cmd/ /src/cmd/
COPY go.mod go.sum /src/
COPY pkg/ /src/pkg/
RUN cd /src && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -extldflags '-zrelro -znow'" -a -installsuffix cgo -o /usr/local/bin/mauzr ./cmd/mauzr

FROM gcr.io/distroless/static:nonroot
USER nobody
COPY --from=build /usr/local/bin/mauzr /usr/local/bin/mauzr
ENTRYPOINT ["/usr/local/bin/mauzr"]
