# Rust version
FROM rust:1.88

# files to copy
COPY ./ ./

RUN cargo build --release

# rodar o producer
CMD [ "./target/release/csv-ingestor" ]
