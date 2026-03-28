use axum::{
    extract::State,
    http::{header, HeaderMap, StatusCode},
    response::IntoResponse,
    Json,
};
use crate::models::{AuthorizeRequest, AuthorizeResponse, User};
use crate::auth::{FirebaseVerifier, IamAuthorizer, AuthenticatedUser};
use std::sync::Arc;
use std::collections::HashMap;

pub struct AppState {
    pub firebase_verifier: Arc<FirebaseVerifier>,
    pub iam_authorizer: Arc<IamAuthorizer>,
    pub role_mapping: HashMap<String, String>,
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

pub async fn authorize(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<AuthorizeRequest>,
) -> Result<impl IntoResponse, (StatusCode, Json<serde_json::Value>)> {
    // 1. Verify token
    let claims = state.firebase_verifier.verify_token(&payload.token).await
        .map_err(|e| (
            StatusCode::UNAUTHORIZED,
            Json(serde_json::json!({ "error": format!("Invalid token: {}", e) }))
        ))?;

    let email = claims.email.ok_or_else(|| (
        StatusCode::BAD_REQUEST,
        Json(serde_json::json!({ "error": "Token missing email" }))
    ))?;

    // 2. Map role alias to GCP role
    let gcp_role = state.role_mapping.get(&payload.required_role)
        .cloned()
        .unwrap_or_else(|| payload.required_role.clone());

    // 3. Check IAM role
    let user_roles = state.iam_authorizer.get_user_iam_roles(&email).await
        .map_err(|e| (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({ "error": format!("IAM check failed: {}", e) }))
        ))?;

    let authorized = user_roles.contains(&gcp_role);

    Ok(Json(AuthorizeResponse {
        authorized,
        uid: claims.sub,
        email,
    }))
}

pub async fn me(
    auth_user: AuthenticatedUser,
) -> impl IntoResponse {
    Json(User {
        uid: auth_user.claims.sub,
        email: auth_user.claims.email.unwrap_or_default(),
        email_verified: auth_user.claims.email_verified.unwrap_or(false),
        roles: auth_user.roles,
    })
}

pub async fn admin_only(
    auth_user: AuthenticatedUser,
) -> impl IntoResponse {
    (StatusCode::OK, format!("Welcome Admin, {}!", auth_user.claims.email.unwrap_or_default()))
}

pub async fn editor_only(
    auth_user: AuthenticatedUser,
) -> impl IntoResponse {
    (StatusCode::OK, format!("Welcome Editor, {}!", auth_user.claims.email.unwrap_or_default()))
}

pub async fn viewer_only(
    auth_user: AuthenticatedUser,
) -> impl IntoResponse {
    (StatusCode::OK, format!("Welcome Viewer, {}!", auth_user.claims.email.unwrap_or_default()))
}
