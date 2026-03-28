use google_cloud_resourcemanager_v3::client::Projects;
use moka::future::Cache;
use serde::{Deserialize, Serialize};
use std::time::Duration;

#[derive(Debug, Serialize, Deserialize, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum AppRole {
    Viewer,
    Editor,
    Admin,
}

pub struct IamAuthorizer {
    project_id: String,
    project_client: Projects,
    cache: Cache<String, Vec<String>>, // email -> list of IAM roles
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

    pub async fn get_user_iam_roles(&self, email: &str) -> Result<Vec<String>, String> {
        if let Some(roles) = self.cache.get(email).await {
            return Ok(roles);
        }

        // Fetch IAM policy
        let policy = self.project_client.get_iam_policy()
            .set_resource(format!("projects/{}", self.project_id))
            .send()
            .await
            .map_err(|e| format!("Failed to get IAM policy: {}", e))?;

        let user_identity = format!("user:{}", email);
        let mut user_roles = Vec::new();

        for binding in policy.bindings {
            if binding.members.contains(&user_identity) {
                user_roles.push(binding.role.clone());
            }
        }

        self.cache.insert(email.to_string(), user_roles.clone()).await;
        Ok(user_roles)
    }

    pub async fn get_user_app_roles(&self, email: &str) -> Result<Vec<AppRole>, String> {
        let iam_roles = self.get_user_iam_roles(email).await?;
        Ok(Self::map_to_app_roles(&iam_roles))
    }

    pub fn map_to_app_roles(iam_roles: &[String]) -> Vec<AppRole> {
        let mut app_roles = Vec::new();
        
        // Admin requirements
        if iam_roles.iter().any(|r| r == "roles/resourcemanager.projectIamAdmin" || r == "roles/bigquery.admin" || 
                                    r == "roles/looker.admin") {
            app_roles.push(AppRole::Admin);
            app_roles.push(AppRole::Editor);
            app_roles.push(AppRole::Viewer);
        } else if iam_roles.iter().any(|r| r == "roles/looker.developer" || 
                                    r == "roles/bigquery.dataEditor" || r == "roles/bigquery.jobUser") {
            // Editor requirements (note: roles/looker.admin is already covered in Admin)
            app_roles.push(AppRole::Editor);
            app_roles.push(AppRole::Viewer);
        } else if iam_roles.iter().any(|r| r == "roles/looker.accessUser" || r == "roles/bigquery.dataViewer") {
            // Viewer requirements
            app_roles.push(AppRole::Viewer);
        }

        app_roles.sort();
        app_roles.dedup();
        app_roles
    }

    pub fn get_highest_role(app_roles: &[AppRole]) -> Option<AppRole> {
        app_roles.iter().max().copied()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_map_admin_roles() {
        let iam_roles = vec!["roles/resourcemanager.projectIamAdmin".to_string()];
        let app_roles = IamAuthorizer::map_to_app_roles(&iam_roles);
        assert!(app_roles.contains(&AppRole::Admin));
        assert!(app_roles.contains(&AppRole::Editor));
        assert!(app_roles.contains(&AppRole::Viewer));
    }

    #[test]
    fn test_map_editor_roles() {
        let iam_roles = vec!["roles/looker.developer".to_string()];
        let app_roles = IamAuthorizer::map_to_app_roles(&iam_roles);
        assert!(!app_roles.contains(&AppRole::Admin));
        assert!(app_roles.contains(&AppRole::Editor));
        assert!(app_roles.contains(&AppRole::Viewer));
    }

    #[test]
    fn test_map_viewer_roles() {
        let iam_roles = vec!["roles/looker.accessUser".to_string()];
        let app_roles = IamAuthorizer::map_to_app_roles(&iam_roles);
        assert!(!app_roles.contains(&AppRole::Admin));
        assert!(!app_roles.contains(&AppRole::Editor));
        assert!(app_roles.contains(&AppRole::Viewer));
    }

    #[test]
    fn test_highest_role() {
        let roles = vec![AppRole::Viewer, AppRole::Editor];
        assert_eq!(IamAuthorizer::get_highest_role(&roles), Some(AppRole::Editor));
    }
}
