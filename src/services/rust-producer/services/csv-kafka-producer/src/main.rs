mod mods;
mod models;

use actix_web::{get, post, web, App, HttpResponse, HttpServer, Responder};
use mods::csv::CsvManager;
use mods::test::{run_auto_test, run_load_test, LoadTestConfig, LoadTestResult};
use rdkafka::config::ClientConfig;
use rdkafka::producer::{DefaultProducerContext, ThreadedProducer};
use serde::Serialize;
use std::sync::Arc;

#[derive(Debug, Serialize)]
struct HealthResponse {
    status: String,
    csv_lines: usize,
}

struct AppState {
    producer: Arc<ThreadedProducer<DefaultProducerContext>>,
    csv_manager: Arc<CsvManager>,
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        Self {
            producer: self.producer.clone(),
            csv_manager: self.csv_manager.clone(),
        }
    }
}

fn create_producer() -> ThreadedProducer<DefaultProducerContext> {
    let bootstrap_servers =
        std::env::var("KAFKA_BOOTSTRAP_SERVERS").unwrap_or_else(|_| "kafka:9092".to_string());

    ClientConfig::new()
        .set("bootstrap.servers", &bootstrap_servers)
        .set("message.timeout.ms", "5000")
        .set("compression.type", "zstd")
        .set("batch.size", "1048576")
        .set("linger.ms", "5")
        .set("queue.buffering.max.messages", "100000")
        .set("acks", "1")
        .set("request.timeout.ms", "30000")
        .create()
        .expect("Falha ao criar producer")
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
    let csv_filepath =
        std::env::var("CSV_FILEPATH").unwrap_or_else(|_| "/app/data/dados.csv".to_string());

    let csv_manager = Arc::new(CsvManager::new(csv_filepath));

    csv_manager
        .initial_load()
        .await
        .expect("falhou ao carregar csv CSV");

    csv_manager.start_auto_reload().await;

    let producer = Arc::new(create_producer());

    let app_state = AppState {
        producer,
        csv_manager,
    };

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(app_state.clone()))
            .service(health)
            .service(csv_post)
    })
    .bind(("0.0.0.0", 5000))?
    .run()
    .await
}
