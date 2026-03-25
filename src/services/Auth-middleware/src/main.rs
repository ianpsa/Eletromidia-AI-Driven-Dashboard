mod handlers;
mod models;
mod jwt;

use axum::{
    routing::{get, post},
    Router,
};
use handlers::{authenticate, authorize, health, healthz, metrics, admin_only, me, AppState};
use jwt::jwt::JwtManager;
use std::env;
use std::net::SocketAddr;
use std::sync::Arc;
use tower_http::trace::TraceLayer;

#[tokio::main]
async fn main() {
    let _ = dotenvy::dotenv();

    let secret = env::var("JWT_SECRET").unwrap_or_else(|_| "nossakeynaovouharcodar".to_string());
    let jwt_manager = Arc::new(JwtManager::new(secret));

    let app_state = Arc::new(AppState {
        jwt_manager,
    });

    let app = Router::new()
        .route("/health", get(health))
        .route("/healthz", get(healthz))
        .route("/metrics", get(metrics))
        .route("/authenticate", post(authenticate))
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

    println!("Auth Middleware started on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .unwrap();
    
    axum::serve(listener, app)
        .await
        .unwrap();
}
