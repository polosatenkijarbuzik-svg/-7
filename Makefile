build:
    mkdir -p bin
    go build --ldflags '-extldflags "-static"' -o bin/invoicer .
