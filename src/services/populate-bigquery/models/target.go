package models

type Target struct {
    ID              string      `bigquery:"id"`
    AgeID           string      `bigquery:"age_id"`
    GenderID        string      `bigquery:"gender_id"`
    SocialClassID   string      `bigquery:"social_class_id"`
}