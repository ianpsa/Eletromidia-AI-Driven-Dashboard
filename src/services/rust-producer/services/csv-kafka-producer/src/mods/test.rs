use crate::models::CsvRow;
use rdkafka::producer::{BaseRecord, DefaultProducerContext, ThreadedProducer};
use serde::{Deserialize, Serialize};
use std::env;
use std::sync::Arc;
use std::time::Instant;
use tokio::task;

fn load_test_batch_size() -> usize {
    env::var("LOAD_TEST_BATCH_SIZE")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(500)
}

fn kafka_topic() -> String {
    env::var("KAFKA_TOPIC").unwrap_or_else(|_| "csv-topic".to_string())
}

fn kafka_record_key() -> String {
    env::var("KAFKA_RECORD_KEY").unwrap_or_else(|_| "csv-row".to_string())
}

fn auto_taxa_inicial_min() -> usize {
    env::var("LOAD_TEST_AUTO_TAXA_INICIAL_MIN")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(100_000)
}

fn auto_taxa_max_limite() -> usize {
    env::var("LOAD_TEST_AUTO_TAXA_MAX_LIMITE")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(5_000_000)
}

fn auto_erro_threshold() -> f64 {
    env::var("LOAD_TEST_AUTO_ERRO_THRESHOLD")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(0.01)
}

fn auto_taxa_min_reducao() -> usize {
    env::var("LOAD_TEST_AUTO_TAXA_MIN_REDUCAO")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(1000)
}

fn auto_fator_aumento() -> f64 {
    env::var("LOAD_TEST_AUTO_FATOR_AUMENTO")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(2.0)
}

fn auto_fator_reducao() -> f64 {
    env::var("LOAD_TEST_AUTO_FATOR_REDUCAO")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(0.7)
}

#[derive(Debug, Deserialize)]
pub struct LoadTestConfig {
    #[serde(default = "default_rate")]
    pub mensagens_por_segundo: usize,
    pub tempo_envio_segundos: usize,
    #[serde(default)]
    pub log_mensagens: bool,
    #[serde(default)]
    pub auto_test: bool,
}

fn default_rate() -> usize {
    env::var("LOAD_TEST_DEFAULT_RATE")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(100000)
}

#[derive(Debug, Serialize)]
pub struct LoadTestResult {
    pub mensagens_enviadas: usize,
    pub mensagens_com_erro: usize,
    pub tempo_total_segundos: f64,
    pub msgs_por_segundo: f64,
}

#[derive(Debug, Serialize)]
pub struct AutoTestResult {
    pub taxa_maxima_sustentavel: usize,
    pub total_enviado: usize,
    pub tempo_total_segundos: f64,
    pub testes: Vec<TestStep>,
}

#[derive(Debug, Serialize)]
pub struct TestStep {
    pub taxa_teste: usize,
    pub mensagens_enviadas: usize,
    pub mensagens_com_erro: usize,
    pub taxa_erro: f64,
}

pub async fn run_load_test(
    target_rate: usize,
    duration_secs: usize,
    csv_data: Arc<Vec<CsvRow>>,
    producer: Arc<ThreadedProducer<DefaultProducerContext>>,
    log_mensagens: bool,
) -> LoadTestResult {
    let batch_size = load_test_batch_size();
    let topic = kafka_topic();
    let key = kafka_record_key();

    let start = Instant::now();
    let mut mensagens_enviadas = 0;
    let mut mensagens_com_erro = 0;

    let target_interval_us = if target_rate > 0 {
        1_000_000 / target_rate as u64
    } else {
        0
    };
    let batch_interval_us = target_interval_us * batch_size as u64;

    while start.elapsed().as_secs() < duration_secs as u64 {
        let batch_start = Instant::now();

        let handles: Vec<_> = (0..batch_size)
            .map(|_| {
                let data = csv_data.clone();
                let producer = producer.clone();
                let topic = topic.clone();
                let key = key.clone();
                async move {
                    let index = rand::random::<usize>() % data.len();
                    let row = &data[index];

                    let payload = match serde_json::to_string(row) {
                        Ok(p) => {
                            if log_mensagens {
                                println!("Enviando: {}", p);
                            }
                            p
                        }
                        Err(_) => return false,
                    };

                    let record = BaseRecord::to(&topic).payload(&payload).key(&key);
                    match producer.send(record) {
                        Ok(_) => true,
                        Err(_) => false,
                    }
                }
            })
            .map(|fut| task::spawn(fut))
            .collect();

        for result in futures::future::join_all(handles).await {
            match result {
                Ok(true) => mensagens_enviadas += 1,
                Ok(false) => mensagens_com_erro += 1,
                Err(_) => mensagens_com_erro += 1,
            }
        }

        let elapsed_us = batch_start.elapsed().as_micros() as u64;
        if elapsed_us < batch_interval_us {
            tokio::time::sleep(std::time::Duration::from_micros(batch_interval_us - elapsed_us))
                .await;
        }
    }

    let tempo_total = start.elapsed().as_secs_f64();
    let msgs_por_segundo = if tempo_total > 0.0 {
        mensagens_enviadas as f64 / tempo_total
    } else {
        0.0
    };

    LoadTestResult {
        mensagens_enviadas,
        mensagens_com_erro,
        tempo_total_segundos: tempo_total,
        msgs_por_segundo,
    }
}

pub async fn run_auto_test(
    target_rate: usize,
    duration_secs: usize,
    csv_data: Arc<Vec<CsvRow>>,
    producer: Arc<ThreadedProducer<DefaultProducerContext>>,
    log_mensagens: bool,
) -> AutoTestResult {
    let taxa_inicial_min = auto_taxa_inicial_min();
    let taxa_max_limite = auto_taxa_max_limite();
    let erro_threshold = auto_erro_threshold();
    let taxa_min_reducao = auto_taxa_min_reducao();
    let fator_aumento = auto_fator_aumento();
    let fator_reducao = auto_fator_reducao();

    let test_duration = duration_secs.max(2);

    let mut testes = Vec::new();
    let mut taxa_inicial = target_rate.max(taxa_inicial_min);
    let mut taxa_maxima = 0;
    let mut total_enviado = 0;
    let mut tempo_total = 0.0;
    let mut erro_taxa_anterior = 0.0;

    eprintln!("Iniciando teste automatico...");

    loop {
        if taxa_inicial > taxa_max_limite {
            eprintln!("Limite atingido: taxa muito alta");
            break;
        }

        eprintln!("Testando taxa: {} msgs/s", taxa_inicial);

        let result = run_load_test(
            taxa_inicial,
            test_duration,
            csv_data.clone(),
            producer.clone(),
            log_mensagens,
        )
        .await;

        let total = result.mensagens_enviadas + result.mensagens_com_erro;
        let taxa_erro = if total > 0 {
            result.mensagens_com_erro as f64 / total as f64
        } else {
            0.0
        };

        total_enviado += result.mensagens_enviadas;
        tempo_total += result.tempo_total_segundos;

        testes.push(TestStep {
            taxa_teste: taxa_inicial,
            mensagens_enviadas: result.mensagens_enviadas,
            mensagens_com_erro: result.mensagens_com_erro,
            taxa_erro,
        });

        if taxa_erro < erro_threshold {
            if taxa_inicial > taxa_maxima {
                taxa_maxima = taxa_inicial;
            }
            erro_taxa_anterior = taxa_erro;
            taxa_inicial = (taxa_inicial as f64 * fator_aumento) as usize;
        } else {
            if erro_taxa_anterior > 0.0 && taxa_inicial > taxa_min_reducao {
                taxa_inicial = (taxa_inicial as f64 * fator_reducao) as usize;
            } else {
                break;
            }
        }
    }

    AutoTestResult {
        taxa_maxima_sustentavel: taxa_maxima,
        total_enviado,
        tempo_total_segundos: tempo_total,
        testes,
    }
}
