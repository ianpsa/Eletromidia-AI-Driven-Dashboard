use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone, PartialEq)]
#[serde(rename_all = "kebab-case")]
pub enum Role {
    User,
    Administrator,
    EletromidiaEmployee,
}

impl Role {
    pub fn as_str(&self) -> &'static str {
        match self {
            Role::User => "user",
            Role::Administrator => "administrator",
            Role::EletromidiaEmployee => "eletromidia-employee",
        }
    }
}

impl From<String> for Role {
    fn from(s: String) -> Self {
        match s.as_str() {
            "administrator" => Role::Administrator,
            "eletromidia-employee" => Role::EletromidiaEmployee,
            _ => Role::User,
        }
    }
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct User {
    pub id: i32,
    pub email: String,
    pub roles: Vec<Role>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AuthRequest {
    pub email: String,
    pub roles: Vec<Role>, // For simplicity in this dummy auth, we accept roles
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AuthResponse {
    pub token: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AuthorizeRequest {
    pub token: String,
    pub required_role: Role,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AuthorizeResponse {
    pub authorized: bool,
    pub user_id: i32,
    pub email: String,
}
