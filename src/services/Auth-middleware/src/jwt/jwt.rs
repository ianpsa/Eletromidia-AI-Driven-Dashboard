use jsonwebtoken::{encode, decode, Header, Validation, EncodingKey, DecodingKey};
use serde::{Deserialize, Serialize};
use chrono::{Utc, Duration};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claims {
    pub sub: String,
    pub user_id: i32,
    pub email: String,
    pub roles: Vec<String>,
    pub exp: i64,
    pub iat: i64,
}

pub struct JwtManager {
    secret: String,
}

impl JwtManager {
    pub fn new(secret: String) -> Self {
        JwtManager { secret }
    }
    
    pub fn generate_token(&self, user_id: i32, email: String, roles: Vec<String>) 
        -> Result<String, jsonwebtoken::errors::Error> 
    {
        let now = Utc::now();
        let expiration = now + Duration::hours(24);
        
        let claims = Claims {
            sub: user_id.to_string(),
            user_id,
            email,
            roles,
            exp: expiration.timestamp(),
            iat: now.timestamp(),
        };
        
        let key = EncodingKey::from_secret(self.secret.as_ref());
        encode(&Header::default(), &claims, &key)
    }
    
    // decript token
    pub fn verify_token(&self, token: &str) -> Result<Claims, jsonwebtoken::errors::Error> {
        let key = DecodingKey::from_secret(self.secret.as_ref());
        let data = decode::<Claims>(token, &key, &Validation::default())?;
        Ok(data.claims)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_jwt_flow() {
        let secret = "test-key".to_string();
        let manager = JwtManager::new(secret);
        
        // generate token
        let token = manager.generate_token(
            1,
            "user@example.com".to_string(),
            vec!["user".to_string()],
        ).unwrap();
        
        // verify token
        let claims = manager.verify_token(&token).unwrap();
        assert_eq!(claims.user_id, 1);
        assert_eq!(claims.email, "user@example.com");
    }
}
