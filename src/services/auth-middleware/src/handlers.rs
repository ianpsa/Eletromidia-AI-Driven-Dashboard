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
use tracing;

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
        .map_err(|e| {
            tracing::error!(error = %e, "Firebase token verification failed in /authorize");
            (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({ "error": format!("Invalid token: {}", e) }))
            )
        })?;

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
        .map_err(|e| {
            tracing::error!(email = %email, error = %e, "IAM role lookup failed in /authorize");
            (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({ "error": format!("IAM check failed: {}", e) }))
            )
        })?;

    let authorized = user_roles.contains(&gcp_role);
    tracing::info!(email = %email, role = %gcp_role, authorized = authorized, "Authorization check completed");

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

pub async fn validate(
    State(state): State<Arc<AppState>>,
    headers: HeaderMap,
) -> impl IntoResponse {
    let auth_header = headers
        .get(header::AUTHORIZATION)
        .and_then(|h| h.to_str().ok());

    let token_str = match auth_header {
        Some(h) => match h.strip_prefix("Bearer ") {
            Some(t) => t.to_string(),
            None => {
                tracing::warn!("Authorization header present but missing 'Bearer ' prefix");
                return StatusCode::UNAUTHORIZED.into_response();
            }
        },
        None => {
            tracing::warn!("Request to /validate missing Authorization header");
            return StatusCode::UNAUTHORIZED.into_response();
        }
    };

    // Verify Firebase token
    let claims = match state.firebase_verifier.verify_token(&token_str).await {
        Ok(c) => c,
        Err(e) => {
            tracing::error!(error = %e, "Firebase token verification failed in /validate");
            return StatusCode::UNAUTHORIZED.into_response();
        }
    };

    // Extract email from claims
    let email = match claims.email {
        Some(ref e) => e.clone(),
        None => {
            tracing::error!(uid = %claims.sub, "Token has no email claim");
            return StatusCode::UNAUTHORIZED.into_response();
        }
    };

    // Check if user has any app role in GCP IAM
    match state.iam_authorizer.get_user_app_roles(&email).await {
        Ok(roles) if !roles.is_empty() => {
            tracing::info!(email = %email, roles = ?roles, "User validated successfully");
            StatusCode::OK.into_response()
        }
        Ok(_) => {
            tracing::warn!(email = %email, "User authenticated but has no mapped app roles — check IAM role mapping");
            StatusCode::FORBIDDEN.into_response()
        }
        Err(e) => {
            tracing::error!(email = %email, error = %e, "IAM role lookup failed in /validate");
            StatusCode::FORBIDDEN.into_response()
        }
    }
}