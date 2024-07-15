package dbt_cloud

// EnvironmentCategory is a type for the different environment categories
type EnvironmentCategory = string

const (
	// All is the category for all environments
	EnvironmentCategory_All EnvironmentCategory = "all"
	// Development is the category for development environments
	EnvironmentCategory_Development EnvironmentCategory = "development"
	// Staging is the category for staging environments
	EnvironmentCategory_Staging EnvironmentCategory = "staging"
	// Production is the category for production environments
	EnvironmentCategory_Production EnvironmentCategory = "production"
	// Other is the category for other environments
	EnvironmentCategory_Other EnvironmentCategory = "other"
)

// EnvironmentCategories is a list of all possible environment categories
var EnvironmentCategories = []EnvironmentCategory{
	EnvironmentCategory_All,
	EnvironmentCategory_Development,
	EnvironmentCategory_Staging,
	EnvironmentCategory_Production,
	EnvironmentCategory_Other,
}
