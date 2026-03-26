use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct User {
    pub uid: String,
    pub email: String,
    pub email_verified: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AuthorizeRequest {
    pub token: String,
    pub required_role: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AuthorizeResponse {
    pub authorized: bool,
    pub uid: String,
    pub email: String,
}
