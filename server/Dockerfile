FROM --platform=linux/arm64 alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite-libs

WORKDIR /root/

# Copy the binary built outside
COPY main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]