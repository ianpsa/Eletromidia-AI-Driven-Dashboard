use google_cloud_resourcemanager_v3::client::Projects;
use moka::future::Cache;
use std::time::Duration;

pub struct IamAuthorizer {
    project_id: String,
    project_client: Projects,
    cache: Cache<(String, String), bool>, // (email, role) -> is_authorized
}

impl IamAuthorizer {
    pub async fn new(project_id: String) -> Result<Self, String> {
        let project_client = Projects::builder().build().await
            .map_err(|e| format!("Failed to create GCP Projects client: {}", e))?;
        
        let cache = Cache::builder()
            .max_capacity(100)
            .time_to_live(Duration::from_secs(60 * 10)) // Cache for 10 minutes
            .build();

        Ok(Self {
            project_id,
            project_client,
            cache,
        })
    }

    pub async fn check_role(&self, email: &str, required_role: &str) -> Result<bool, String> {
        let key = (email.to_string(), required_role.to_string());
        if let Some(authorized) = self.cache.get(&key).await {
            return Ok(authorized);
        }

        // Fetch IAM policy using builder pattern
        let policy = self.project_client.get_iam_policy()
            .set_resource(format!("projects/{}", self.project_id))
            .send()
            .await
            .map_err(|e| format!("Failed to get IAM policy: {}", e))?;

        let user_identity = format!("user:{}", email);
        let mut is_authorized = false;

        for binding in policy.bindings {
            if binding.role == required_role {
                if binding.members.contains(&user_identity) {
                    is_authorized = true;
                    break;
                }
            }
        }

        self.cache.insert(key, is_authorized).await;
        Ok(is_authorized)
    }
}
