mod mods;
mod models;

use actix_web::{get, post, web, App, HttpResponse, HttpServer, Responder};
use std::env;
use mods::csv::CsvManager;
use mods::test::{run_auto_test, run_load_test, LoadTestConfig, LoadTestResult};
use prometheus_client::encoding::EncodeLabelSet;
use prometheus_client::metrics::counter::Counter;
use prometheus_client::metrics::gauge::Gauge;
use prometheus_client::registry::Registry;
use rdkafka::config::ClientConfig;
use rdkafka::producer::{DefaultProducerContext, ThreadedProducer};
use serde::Serialize;
use std::sync::Arc;
use parking_lot::RwLock;

#[derive(Debug, Clone, EncodeLabelSet)]
struct Labels {
    service: String,
}

#[derive(Debug, Serialize)]
struct HealthResponse {
    status: String,
    csv_lines: usize,
}

struct AppState {
    producer: Arc<ThreadedProducer<DefaultProducerContext>>,
    csv_manager: Arc<CsvManager>,
    metrics_registry: Arc<RwLock<Registry>>,
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            producer: self.producer.clone(),
            csv_manager: self.csv_manager.clone(),
            metrics_registry: self.metrics_registry.clone(),
        }
    }
}

fn create_metrics_registry() -> Registry {
    let registry = Registry::default();
    registry
}

fn create_producer() -> ThreadedProducer<DefaultProducerContext> {
    let bootstrap_servers = env::var("KAFKA_BOOTSTRAP_SERVERS").unwrap_or_else(|_| "kafka:9092".to_string());
    let message_timeout_ms = env::var("KAFKA_MESSAGE_TIMEOUT_MS").unwrap_or_else(|_| "5000".to_string());
    let compression_type = env::var("KAFKA_COMPRESSION_TYPE").unwrap_or_else(|_| "zstd".to_string());
    let batch_size = env::var("KAFKA_BATCH_SIZE").unwrap_or_else(|_| "1048576".to_string());
    let linger_ms = env::var("KAFKA_LINGER_MS").unwrap_or_else(|_| "5".to_string());
    let queue_buffering_max = env::var("KAFKA_QUEUE_BUFFERING_MAX_MESSAGES").unwrap_or_else(|_| "100000".to_string());
    let acks = env::var("KAFKA_ACKS").unwrap_or_else(|_| "1".to_string());
    let request_timeout_ms = env::var("KAFKA_REQUEST_TIMEOUT_MS").unwrap_or_else(|_| "30000".to_string());

    ClientConfig::new()
        .set("bootstrap.servers", &bootstrap_servers)
        .set("message.timeout.ms", &message_timeout_ms)
        .set("compression.type", &compression_type)
        .set("batch.size", &batch_size)
        .set("linger.ms", &linger_ms)
        .set("queue.buffering.max.messages", &queue_buffering_max)
        .set("acks", &acks)
        .set("request.timeout.ms", &request_timeout_ms)
        .create()
        .expect("Falha ao criar producer")
}

#[get("/metrics")]
async fn metrics(state: web::Data<AppState>) -> impl Responder {
    let registry = state.metrics_registry.read();
    let mut buf = Vec::new();
    let encoder = prometheus_client::encoding::TextEncoder::new();
    if encoder.encode(&registry.collect(), &mut buf).is_ok() {
        HttpResponse::Ok()
            .content_type("application/openmetrics-text; version=1.0.0; charset=utf-8")
            .body(String::from_utf8(buf).unwrap_or_default())
    } else {
        HttpResponse::InternalServerError().body("Failed to encode metrics")
    }
}

#[get("/healthz")]
async fn healthz() -> impl Responder {
    HttpResponse::Ok().body("ok")
}

#[get("/health")]
async fn health(state: web::Data<AppState>) -> impl Responder {
    let csv_lines = {
        let cache = state.csv_manager.cache.read().await;
        cache.len()
    };

    HttpResponse::Ok().json(HealthResponse {
        status: "ok".to_string(),
        csv_lines,
    })
}

#[post("/csv")]
async fn csv_post(config: web::Json<LoadTestConfig>, state: web::Data<AppState>) -> actix_web::Result<impl Responder> {
    let csv_data = {
        let cache = state.csv_manager.cache.read().await;
        if cache.is_empty() {
            let result = LoadTestResult {
                mensagens_enviadas: 0,
                mensagens_com_erro: 0,
                tempo_total_segundos: 0.0,
                msgs_por_segundo: 0.0,
            };
            return Ok(HttpResponse::Ok().json(result));
        }
        cache.clone()
    };

    let producer = state.producer.clone();

    if config.auto_test {
        let result = run_auto_test(
            config.mensagens_por_segundo,
            config.tempo_envio_segundos,
            csv_data,
            producer,
            config.log_mensagens,
        )
        .await;
        return Ok(HttpResponse::Ok().json(result));
    }

    let result = run_load_test(
        config.mensagens_por_segundo,
        config.tempo_envio_segundos,
        csv_data,
        producer,
        config.log_mensagens,
    )
    .await;

    Ok(HttpResponse::Ok().json(result))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let _ = dotenvy::dotenv();

    let csv_filepath = env::var("CSV_FILEPATH").unwrap_or_else(|_| "/app/data/dados.csv".to_string());
    let reload_interval_secs: u64 = env::var("CSV_RELOAD_INTERVAL_SECS")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(86400);

    let csv_manager = Arc::new(CsvManager::new(csv_filepath, reload_interval_secs));

    csv_manager
        .initial_load()
        .await
        .expect("falhou ao carregar csv CSV");

    csv_manager.start_auto_reload().await;

    let producer = Arc::new(create_producer());

    let metrics_registry = Arc::new(RwLock::new(create_metrics_registry()));

    let app_state = AppState {
        producer,
        csv_manager,
        metrics_registry,
    };

    let http_host = env::var("HTTP_HOST").unwrap_or_else(|_| "0.0.0.0".to_string());
    let http_port: u16 = env::var("HTTP_PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(5000);

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(app_state.clone()))
            .service(metrics)
            .service(healthz)
            .service(health)
            .service(csv_post)
    })
    .bind((http_host, http_port))?
    .run()
    .await
}
