use axum::{
    extract::FromRequestParts,
    http::{request::Parts, StatusCode, header::AUTHORIZATION},
    response::{IntoResponse, Response},
};
use crate::auth::FirebaseClaims;
use crate::handlers::AppState;
use serde::Serialize;
use std::sync::Arc;

#[derive(Debug, Serialize)]
pub enum AuthError {
    MissingToken,
    InvalidToken(String),
    InsufficientPermissions,
    InternalError(String),
}

impl IntoResponse for AuthError {
    fn into_response(self) -> Response {
        let (status, message) = match self {
            AuthError::MissingToken => (
                StatusCode::UNAUTHORIZED,
                "Authorization token missing".to_string(),
            ),
            AuthError::InvalidToken(e) => (
                StatusCode::UNAUTHORIZED,
                format!("Invalid authorization token: {}", e),
            ),
            AuthError::InsufficientPermissions => (
                StatusCode::FORBIDDEN,
                "Insufficient permissions for this resource".to_string(),
            ),
            AuthError::InternalError(e) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                format!("Internal server error: {}", e),
            ),
        };

        (status, message).into_response()
    }
}

pub struct AuthToken(pub FirebaseClaims);

impl FromRequestParts<Arc<AppState>> for AuthToken 
{
    type Rejection = AuthError;
    
    async fn from_request_parts(parts: &mut Parts, state: &Arc<AppState>) -> Result<Self, Self::Rejection> {
        let auth_header = parts
            .headers
            .get(AUTHORIZATION)
            .and_then(|h| h.to_str().ok())
            .ok_or(AuthError::MissingToken)?;
        
        let token_str = auth_header
            .strip_prefix("Bearer ")
            .ok_or(AuthError::InvalidToken("Expected Bearer token".to_string()))?
            .to_string(); // Clone to release borrow on parts
        
        // state is already Arc<AppState>
        let app_state = state;

        let claims = app_state.firebase_verifier
            .verify_token(&token_str)
            .await
            .map_err(AuthError::InvalidToken)?;
        
        Ok(AuthToken(claims))
    }
}
