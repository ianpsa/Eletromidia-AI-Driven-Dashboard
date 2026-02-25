use csv::ReaderBuilder;
use rand::{self, RngExt};
use std::error::Error;

async fn read_csv(rand: usize, filepath: &str) -> Result<Vec<String>, Box<dyn Error>> {

    let mut rdr = ReaderBuilder::new().from_path(filepath)?;
    
    if let Some(result) = rdr.records().nth(rand) {
        let record = result?;
        let values: Vec<String> = record.iter().map(|s| s.to_string()).collect();
        return Ok(values);
    }
    
    Err("Index fora de escopo".into())
}

pub async fn take_csv_line() -> Result<Vec<String>, Box<dyn Error>> {
  
    let filepath = "../../services/csv-ingestor/data/dados.csv";
    
    let mut rdr = ReaderBuilder::new().from_path(&filepath)?;
    
    let total_records = rdr.records().count();
    
    if total_records == 0 {
      return Err("CSV está vazio".into());
    }
    
    println!("O CSV tem {} linhas", total_records);
    
    
    let rng = rand::rng().random_range(0..total_records);
    let rand_record = read_csv(rng, filepath).await?;
    
    println!("{:?} peguei esta linha", rand_record);
    
    Ok(rand_record)
}