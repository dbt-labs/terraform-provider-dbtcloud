variable "project_id" {
  description = "The ID of the dbt Cloud project"
  type        = number
}

variable "num_threads" {
  description = "Number of threads to use"
  type        = string
}

variable "private_key_id" {
  description = "Private Key ID for the Service Account"
  type        = bool
}

variable "private_key" {
  description = "Private Key for the Service Account"
  type        = bool
}

variable "client_email" {
  description = "Service Account email"
  type        = bool
}

variable "client_id" {
  description = "Client ID of the Service Account"
  type        = bool
}

variable "auth_uri" {
  description = "Auth URI for the Service Account"
  type        = bool
}

variable "token_uri" {
  description = "Token URI for the Service Account"
  type        = bool
}

variable "auth_provider_x509_cert_url" {
  description = "Auth Provider X509 Cert URL for the Service Account"
  type        = bool
}

variable "client_x509_cert_url" {
  description = "Client X509 Cert URL for the Service Account"
  type        = bool
}