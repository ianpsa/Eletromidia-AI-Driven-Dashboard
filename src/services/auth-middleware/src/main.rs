mod handlers;
mod models;
mod auth;

use axum::{
    routing::{get, post},
    Router,
};
use handlers::{authorize, health, healthz, metrics, admin_only, me, AppState};
use auth::{FirebaseVerifier, IamAuthorizer};
use std::env;
use std::net::SocketAddr;
use std::sync::Arc;
use std::collections::HashMap;
use tower_http::trace::TraceLayer;

#[tokio::main]
async fn main() {
    let _ = dotenvy::dotenv();

    let firebase_project_id = env::var("FIREBASE_PROJECT_ID")
        .expect("FIREBASE_PROJECT_ID must be set");
    
    let gcp_project_id = env::var("GCP_PROJECT_ID")
        .expect("GCP_PROJECT_ID must be set");

    let firebase_verifier = Arc::new(FirebaseVerifier::new(firebase_project_id));
    
    let iam_authorizer = Arc::new(IamAuthorizer::new(gcp_project_id).await
        .expect("Failed to initialize IAM authorizer"));

    // Load role mapping from environment variables starting with ROLE_
    let mut role_mapping = HashMap::new();
    for (key, value) in env::vars() {
        if let Some(alias) = key.strip_prefix("ROLE_") {
            role_mapping.insert(alias.to_lowercase(), value);
        }
    }

    let app_state = Arc::new(AppState {
        firebase_verifier,
        iam_authorizer,
        role_mapping,
    });

    let app = Router::new()
        .route("/health", get(health))
        .route("/healthz", get(healthz))
        .route("/metrics", get(metrics))
        .route("/authorize", post(authorize))
        .route("/me", get(me))
        .route("/admin", get(admin_only))
        .layer(TraceLayer::new_for_http())
        .with_state(app_state);

    let host = env::var("HTTP_HOST").unwrap_or_else(|_| "0.0.0.0".to_string());
    let port: u16 = env::var("HTTP_PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(5000);

    let addr: SocketAddr = format!("{}:{}", host, port)
        .parse()
        .expect("Invalid address format");

    println!("Auth Middleware with Firebase & GCP IAM started on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .unwrap();
    
    axum::serve(listener, app)
        .await
        .unwrap();
}
