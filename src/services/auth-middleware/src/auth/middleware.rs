use axum::{
    extract::FromRequestParts,
    http::{request::Parts, StatusCode, header::AUTHORIZATION},
    response::{IntoResponse, Response},
    middleware::Next,
    extract::State,
};
use crate::auth::{FirebaseClaims, AppRole};
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

#[derive(Clone, Debug)]
pub struct AuthenticatedUser {
    pub claims: FirebaseClaims,
    pub roles: Vec<AppRole>,
}

impl FromRequestParts<Arc<AppState>> for AuthenticatedUser 
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
        
        let app_state = state;

        let claims = app_state.firebase_verifier
            .verify_token(&token_str)
            .await
            .map_err(AuthError::InvalidToken)?;
        
        let email = claims.email.as_ref().ok_or(AuthError::InvalidToken("Missing email in token".to_string()))?;
        
        let roles = app_state.iam_authorizer.get_user_app_roles(email)
            .await
            .map_err(AuthError::InternalError)?;
        
        Ok(AuthenticatedUser { claims, roles })
    }
}

use axum::extract::Request;

// Middleware that requires a specific role
pub async fn require_role(
    State(_state): State<Arc<AppState>>,
    auth_user: AuthenticatedUser,
    request: Request,
    next: Next,
    required_role: AppRole,
) -> Result<Response, AuthError> {
    if auth_user.roles.contains(&required_role) {
        Ok(next.run(request).await)
    } else {
        Err(AuthError::InsufficientPermissions)
    }
}

pub async fn require_admin(
    state: State<Arc<AppState>>,
    auth: AuthenticatedUser,
    req: Request,
    next: Next,
) -> Result<Response, AuthError> {
    require_role(state, auth, req, next, AppRole::Admin).await
}

pub async fn require_editor(
    state: State<Arc<AppState>>,
    auth: AuthenticatedUser,
    req: Request,
    next: Next,
) -> Result<Response, AuthError> {
    require_role(state, auth, req, next, AppRole::Editor).await
}

pub async fn require_viewer(
    state: State<Arc<AppState>>,
    auth: AuthenticatedUser,
    req: Request,
    next: Next,
) -> Result<Response, AuthError> {
    require_role(state, auth, req, next, AppRole::Viewer).await
}
