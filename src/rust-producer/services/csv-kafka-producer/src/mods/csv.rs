use crate::models::{CsvRow, TargetData};
use csv::ReaderBuilder;
use std::error::Error;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;

pub type CsvCache = Arc<RwLock<Arc<Vec<CsvRow>>>>;

pub struct CsvManager {
    pub cache: CsvCache,
    filepath: String,
    reload_interval: Duration,
}

impl CsvManager {
    pub fn new(filepath: String) -> Self {
        Self {
            cache: Arc::new(RwLock::new(Arc::new(Vec::new()))),
            filepath,
            reload_interval: Duration::from_secs(86400),
        }
    }

    pub async fn start_auto_reload(&self) {
        let cache = self.cache.clone();
        let filepath = self.filepath.clone();
        let interval = self.reload_interval;

        tokio::spawn(async move {
            loop {
                tokio::time::sleep(interval).await;
                match Self::load_csv(&filepath) {
                    Ok(records) => {
                        let mut write = cache.write().await;
                        *write = records;
                        println!("CSV recarregado com sucesso");
                    }
                    Err(e) => {
                        eprintln!("Erro ao recarregar CSV: {}", e);
                    }
                }
            }
        });
    }

    pub async fn initial_load(&self) -> Result<(), Box<dyn Error + Send + Sync>> {
        let records = Self::load_csv(&self.filepath)?;
        let mut cache = self.cache.write().await;
        *cache = records;
        println!("CSV carregado inicialmente");
        Ok(())
    }

    fn load_csv(filepath: &str) -> Result<Arc<Vec<CsvRow>>, Box<dyn Error + Send + Sync>> {
        let mut rdr = ReaderBuilder::new().from_path(filepath)?;
        let mut records = Vec::new();

        for result in rdr.records() {
            let csv_lines = match result {
                Ok(r) => r,
                Err(_) => continue,
            };

            if csv_lines.len() < 10 {
                continue;
            }

            let impression_hour: usize = csv_lines[0].parse().unwrap_or(0);
            let location_id: i64 = csv_lines[1].parse().unwrap_or(0);
            let uniques: f64 = csv_lines[2].parse().unwrap_or(0.0);
            let latitude: f64 = csv_lines[3].parse().unwrap_or(0.0);
            let longitude: f64 = csv_lines[4].parse().unwrap_or(0.0);
            let numero: Option<isize> = csv_lines[8].parse().ok();

            let target: Option<TargetData> =
                serde_json::from_str(&csv_lines[9].replace("'", "\"")).ok();

            records.push(CsvRow {
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
            });
        }

        if records.is_empty() {
            return Err("CSV está vazio :/".into());
        }

        println!("CSV carregado com {} linhas", records.len());
        Ok(Arc::new(records))
    }
}
