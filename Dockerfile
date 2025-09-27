    #TODO: Get this working, windows builds aren't working yet :(

    # Stage 1: Build the Go application for Windows
    FROM golang:latest AS builder

    # Set environment variables for cross-compilation to Windows
    ENV CGO_ENABLED=0
    ENV GOOS=windows
    ENV GOARCH=amd64 # or 386 for 32-bit Windows

    WORKDIR /app

    COPY . .

    # Build the Go application
    RUN go build -o lomo.exe .

    # Stage 2: Create a lightweight image containing only the Windows executable
    FROM scratch

    # Copy the built executable from the builder stage
    COPY --from=builder /app/myapp.exe /bin/lomo_win_x64.exe

    # Define the command to run the executable
    #CMD ["/lomo.exe"]
