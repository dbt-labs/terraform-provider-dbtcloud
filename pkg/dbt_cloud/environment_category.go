package dbt_cloud

// EnvironmentCategory is a type for the different environment categories
type EnvironmentCategory = string

const (
	// All is the category for all environments
	All EnvironmentCategory = "all"
	// Development is the category for development environments
	Development EnvironmentCategory = "development"
	// Staging is the category for staging environments
	Staging EnvironmentCategory = "staging"
	// Production is the category for production environments
	Production EnvironmentCategory = "production"
	// Other is the category for other environments
	Other EnvironmentCategory = "other"
)

// EnvironmentCategories is a list of all possible environment categories
var EnvironmentCategories = []EnvironmentCategory{
	All,
	Development,
	Staging,
	Production,
	Other,
}
