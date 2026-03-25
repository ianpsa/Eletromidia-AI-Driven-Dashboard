use axum::{
    extract::State,
    http::{header, HeaderMap, StatusCode},
    response::IntoResponse,
    Json,
};
use crate::models::{AuthRequest, AuthResponse, AuthorizeRequest, AuthorizeResponse, Role, User};
use crate::jwt::{JwtManager, AuthError};
use std::sync::Arc;

pub struct AppState {
    pub jwt_manager: Arc<JwtManager>,
}

pub async fn health() -> impl IntoResponse {
    Json(serde_json::json!({ "status": "ok" }))
}

pub async fn healthz() -> impl IntoResponse {
    "ok"
}

pub async fn metrics() -> impl IntoResponse {
    let mut headers = HeaderMap::new();
    headers.insert(header::CONTENT_TYPE, "text/plain; version=0.0.4".parse().unwrap());
    
    (
        headers,
        "# HELP auth_middleware_up Prometheus exporter for auth-middleware\n# TYPE auth_middleware_up gauge\nauth_middleware_up 1\n"
    )
}

pub async fn authenticate(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<AuthRequest>,
) -> Result<impl IntoResponse, (StatusCode, Json<serde_json::Value>)> {

    
    let user_id = 123;
    let roles_strings: Vec<String> = payload.roles.iter().map(|r| r.as_str().to_string()).collect();

    let token = state.jwt_manager.generate_token(
        user_id,
        payload.email,
        roles_strings
    ).map_err(|e| (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(serde_json::json!({ "error": format!("Failed to generate token: {}", e) }))
    ))?;

    Ok(Json(AuthResponse { token }))
}

pub async fn authorize(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<AuthorizeRequest>,
) -> Result<impl IntoResponse, (StatusCode, Json<serde_json::Value>)> {
    let claims = state.jwt_manager.verify_token(&payload.token).map_err(|_| (
        StatusCode::UNAUTHORIZED,
        Json(serde_json::json!({ "error": "Invalid token" }))
    ))?;

    let roles_as_enums: Vec<Role> = claims.roles.into_iter().map(Role::from).collect();
    
    let is_authorized = roles_as_enums.contains(&payload.required_role);

    Ok(Json(AuthorizeResponse {
        authorized: is_authorized,
        user_id: claims.user_id,
        email: claims.email,
    }))
}

pub async fn me(
    crate::jwt::AuthToken(claims): crate::jwt::AuthToken,
) -> impl IntoResponse {
    let roles: Vec<Role> = claims.roles.into_iter().map(Role::from).collect();
    
    Json(User {
        id: claims.user_id,
        email: claims.email,
        roles,
    })
}

pub async fn admin_only(
    crate::jwt::AuthToken(claims): crate::jwt::AuthToken,
) -> Result<impl IntoResponse, AuthError> {
    let roles: Vec<Role> = claims.roles.into_iter().map(Role::from).collect();

    if roles.contains(&Role::Administrator) {
        Ok((StatusCode::OK, format!("Welcome Admin, {}!", claims.email)))
    } else {
        Err(AuthError::InsufficientPermissions)
    }
}
