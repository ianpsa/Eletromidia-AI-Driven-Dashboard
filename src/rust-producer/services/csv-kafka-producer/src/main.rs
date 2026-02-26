use actix_web::{post, web, App, HttpResponse, HttpServer, Responder};
use rdkafka::config::ClientConfig;
use rdkafka::producer::{FutureProducer, FutureRecord};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::{Duration, Instant};

mod mods;

type TargetData = HashMap<String, HashMap<String, f64>>;

#[derive(Debug, Serialize, Deserialize, Clone)]
struct CsvRow {
    impression_hour: usize,
    location_id: i64,
    uniques: f64,
    latitude: f64,
    longitude: f64,
    uf_estado: String,
    cidade: String,
    endereco: String,
    numero: isize,
    target: TargetData,
}

#[derive(Debug, Deserialize)]
struct LoadTestConfig {
    mensagens_por_segundo: usize,
    tempo_envio_segundos: usize,
}

#[derive(Debug, Serialize)]
struct LoadTestResult {
    mensagens_enviadas: usize,
    mensagens_com_erro: usize,
    tempo_total_segundos: f64,
}

fn create_producer() -> FutureProducer {
    ClientConfig::new()
        .set("bootstrap.servers", "localhost:9092")
        .set("message.timeout.ms", "5000")
        .create()
        .expect("Falha ao criar producer")
}

async fn build_csv_row() -> Result<CsvRow, String> {
    let csv_lines = mods::csv::take_csv_line().await.map_err(|e| format!("Falha ao ler CSV: {}", e))?;

    if csv_lines.len() < 10 {
        return Err("Linha CSV incompleta".to_string());
    }

    let impression_hour: usize = csv_lines[0].parse().map_err(|_| "Erro no parsing do impression_hour")?;
    let location_id: i64 = csv_lines[1].parse().map_err(|_| "Erro no parsing do location_id")?;
    let uniques: f64 = csv_lines[2].parse().map_err(|_| "Erro no parsing do uniques")?;
    let latitude: f64 = csv_lines[3].parse().map_err(|_| "Erro no parsing do latitude")?;
    let longitude: f64 = csv_lines[4].parse().map_err(|_| "Erro no parsing do longitude")?;
    let numero: isize = csv_lines[8].parse().map_err(|_| "Erro no parsing do numero")?;
    let target: TargetData = serde_json::from_str(&csv_lines[9].replace("'", "\""))
        .map_err(|_| "Erro no parsing do target")?;

    Ok(CsvRow {
        impression_hour,
        location_id,
        uniques,
        latitude,
        longitude,
        uf_estado: csv_lines[5].to_string(),
        cidade: csv_lines[6].to_string(),
        endereco: csv_lines[7].to_string(),
        numero,
        target,
    })
}

#[post("/csv")]
async fn csv_post(config: web::Json<LoadTestConfig>) -> impl Responder {
    let producer = create_producer();
    let interval = Duration::from_micros(1_000_000 / config.mensagens_por_segundo as u64);
    let total_duration = Duration::from_secs(config.tempo_envio_segundos as u64);

    let start = Instant::now();
    let mut mensagens_enviadas = 0;
    let mut mensagens_com_erro = 0;

    while start.elapsed() < total_duration {
        let loop_start = Instant::now();

        let row = match build_csv_row().await {
            Ok(r) => r,
            Err(e) => {
                eprintln!("Erro ao construir mensagem: {}", e);
                mensagens_com_erro += 1;
                continue;
            }
        };

        let payload = match serde_json::to_string(&row) {
            Ok(p) => p,
            Err(e) => {
                eprintln!("Erro ao serializar: {}", e);
                mensagens_com_erro += 1;
                continue;
            }
        };

        if let Err((e, _)) = producer
            .send(
                FutureRecord::to("csv-topic")
                    .payload(&payload)
                    .key("csv-row"),
                Duration::from_secs(5),
            )
            .await
        {
            eprintln!("Erro ao enviar para Kafka: {}", e);
            mensagens_com_erro += 1;
        } else {
            mensagens_enviadas += 1;
        }

        let elapsed = loop_start.elapsed();
        if elapsed < interval {
            tokio::time::sleep(interval - elapsed).await;
        }
    }

    let result = LoadTestResult {
        mensagens_enviadas,
        mensagens_com_erro,
        tempo_total_segundos: start.elapsed().as_secs_f64(),
    };

    HttpResponse::Ok().json(result)
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| App::new().service(csv_post))
        .bind(("127.0.0.1", 5000))?
        .run()
        .await
}
