
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

pub type TargetData = HashMap<String, HashMap<String, f64>>;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CsvRow {
    pub impression_hour: usize,
    pub location_id: i64,
    pub uniques: f64,
    pub latitude: String,
    pub longitude: String,
    pub uf_estado: String,
    pub cidade: String,
    pub endereco: String,
    pub numero: Option<isize>,
    pub target: Option<TargetData>,
}
