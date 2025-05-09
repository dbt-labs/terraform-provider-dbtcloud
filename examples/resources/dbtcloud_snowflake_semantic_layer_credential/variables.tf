variable "project_id" {
  description = "The ID of the dbt Cloud project"
  type        = number
}

# Snowflake connection details
variable "database" {
  description = "The Snowflake database name"
  type        = string
}

variable "schema" {
  description = "The Snowflake schema name"
  type        = string
}

variable "warehouse" {
  description = "The Snowflake warehouse name"
  type        = string
}

variable "role" {
  description = "The Snowflake role name"
  type        = string
}

# Password authentication variables
variable "user" {
  description = "The Snowflake username (for password auth)"
  type        = string
  default     = ""
}

variable "password" {
  description = "The Snowflake password (for password auth)"
  type        = string
  default     = ""
  sensitive   = true
}

# Key pair authentication variables
variable "private_key" {
  description = "The private key for Snowflake authentication (for keypair auth)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "private_key_passphrase" {
  description = "The passphrase for the private key (for keypair auth)"
  type        = string
  default     = ""
  sensitive   = true
} 