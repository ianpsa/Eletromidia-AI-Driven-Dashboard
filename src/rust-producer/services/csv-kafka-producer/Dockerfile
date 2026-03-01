# Build stage
FROM rust:1.88 as builder

WORKDIR /build

COPY Cargo.toml ./
COPY src ./src

RUN apt-get update && apt-get install -y cmake libssl-dev pkg-config libzstd-dev clang && \
    cargo build --release

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y libssl3 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/target/release/csv-kafka-producer /app/

COPY data /app/data

EXPOSE 5000

CMD ["/app/csv-kafka-producer"]
