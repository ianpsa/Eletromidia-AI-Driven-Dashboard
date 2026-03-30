use jsonwebtoken::{decode, decode_header, DecodingKey, Validation, Algorithm};
use serde::{Deserialize, Serialize};
use moka::future::Cache;
use std::time::Duration;
use reqwest::Client;

const GOOGLE_CERTS_URL: &str = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com";

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct FirebaseClaims {
    pub aud: String,
    pub exp: i64,
    pub iat: i64,
    pub iss: String,
    pub sub: String,
    pub email: Option<String>,
    pub email_verified: Option<bool>,
    pub firebase: FirebaseData,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct FirebaseData {
    pub identities: Option<serde_json::Value>,
    pub sign_in_provider: String,
}

pub struct FirebaseVerifier {
    project_id: String,
    client: Client,
    key_cache: Cache<String, DecodingKey>,
}

impl FirebaseVerifier {
    pub fn new(project_id: String) -> Self {
        let key_cache = Cache::builder()
            .max_capacity(10)
            .time_to_live(Duration::from_secs(60 * 60 * 6)) // Cache for 6 hours
            .build();
        
        Self {
            project_id,
            client: Client::new(),
            key_cache,
        }
    }

    async fn get_decoding_key(&self, kid: &str) -> Result<DecodingKey, String> {
        if let Some(key) = self.key_cache.get(kid).await {
            return Ok(key);
        }

        // Fetch fresh keys from Google
        let response = self.client.get(GOOGLE_CERTS_URL)
            .send()
            .await
            .map_err(|e| format!("Failed to fetch Google certs: {}", e))?;

        let certs: serde_json::Value = response.json()
            .await
            .map_err(|e| format!("Failed to parse Google certs: {}", e))?;

        let cert_pem = certs.get(kid)
            .and_then(|v| v.as_str())
            .ok_or_else(|| format!("Key ID {} not found in Google certs", kid))?;

        let decoding_key = DecodingKey::from_rsa_pem(cert_pem.as_bytes())
            .map_err(|e| format!("Failed to create decoding key: {}", e))?;

        self.key_cache.insert(kid.to_string(), decoding_key.clone()).await;
        Ok(decoding_key)
    }

    pub async fn verify_token(&self, token: &str) -> Result<FirebaseClaims, String> {
        let header = decode_header(token)
            .map_err(|e| format!("Failed to decode token header: {}", e))?;

        let kid = header.kid
            .ok_or_else(|| "Missing Key ID (kid) in token header".to_string())?;

        let decoding_key = self.get_decoding_key(&kid).await?;

        let mut validation = Validation::new(Algorithm::RS256);
        validation.set_audience(&[self.project_id.clone()]);
        validation.set_issuer(&[format!("https://securetoken.google.com/{}", self.project_id)]);

        let token_data = decode::<FirebaseClaims>(token, &decoding_key, &validation)
            .map_err(|e| format!("Token validation failed: {}", e))?;

        Ok(token_data.claims)
    }
}
