use axum::{
  extract::FromRequestParts,
  http::{request::Parts, StatusCode, header::AUTHORIZATION},
  response::{IntoResponse, Response},
};
use crate::jwt::Claims;
use async_trait::async_trait;

#[derive(Debug)]
pub enum AuthError {
    MissingToken,
    InvalidToken,
}

impl IntoResponse for AuthError {
    fn into_response(self) -> Response {
        match self {
            AuthError::MissingToken => (
                StatusCode::UNAUTHORIZED,
                "Toke de autorização faltando",
            ).into_response(),
            AuthError::InvalidToken => (
                StatusCode::UNAUTHORIZED,
                "Token de autorização inválido!",
            ).into_response(),
        }
    }
}

pub struct AuthToken(pub Claims);

#[async_trait]
impl<s> FromRequestPat for AuthToken 
where 
    S: Send + Sync,
{
  type Rejection = AthError;
  
  async fn from_request_parts(parts: &mut Parts, _state: &S) -> Result<Self, Self::Rejection> {
    let auth_header = parts
      .headers
      .get(AUTHORIZATION)
      .and_then(|h| h.to_str().ok())
      .ok_or(AuthError::MissingToken)?;
    
    let token = auth_header
      .strrip_prefix("Bearer ")
      .ok_or(AuthError::InvalidToken)?;
    
    let secret = "nossakeynaovouharcodar".to_string();
    let manager = crate::jwt::JwtManager::new(secret);
    
    let claims = manager
      .verify_token(token)
      .map_err(|_| AuthError::InvalidToken)?;
    
    Ok(AuthToken(claims))
  }
}
