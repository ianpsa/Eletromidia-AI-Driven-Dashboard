use axum::{
    extract::FromRequestParts,
    http::{request::Parts, StatusCode, header::AUTHORIZATION},
    response::{IntoResponse, Response},
};
use crate::jwt::Claims;
use serde::Serialize;
use std::env;

#[derive(Debug, Serialize)]
pub enum AuthError {
    MissingToken,
    InvalidToken,
    InsufficientPermissions,
}

impl IntoResponse for AuthError {
    fn into_response(self) -> Response {
        let (status, message) = match self {
            AuthError::MissingToken => (
                StatusCode::UNAUTHORIZED,
                "Authorization token missing",
            ),
            AuthError::InvalidToken => (
                StatusCode::UNAUTHORIZED,
                "Invalid authorization token!",
            ),
            AuthError::InsufficientPermissions => (
                StatusCode::FORBIDDEN,
                "Insufficient permissions for this resource",
            ),
        };

        (status, message).into_response()
    }
}

pub struct AuthToken(pub Claims);

impl<S> FromRequestParts<S> for AuthToken 
where 
    S: Send + Sync,
{
    type Rejection = AuthError;
    
    async fn from_request_parts(parts: &mut Parts, _state: &S) -> Result<Self, Self::Rejection> {
        let auth_header = parts
            .headers
            .get(AUTHORIZATION)
            .and_then(|h| h.to_str().ok())
            .ok_or(AuthError::MissingToken)?;
        
        let token = auth_header
            .strip_prefix("Bearer ")
            .ok_or(AuthError::InvalidToken)?;
        
        let secret = env::var("JWT_SECRET").unwrap_or_else(|_| "nossakeynaovouharcodar".to_string());
        let manager = crate::jwt::JwtManager::new(secret);
        
        let claims = manager
            .verify_token(token)
            .map_err(|_| AuthError::InvalidToken)?;
        
        Ok(AuthToken(claims))
    }
}
