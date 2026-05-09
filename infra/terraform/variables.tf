variable "region" {
  type        = string
  default     = "lon1"
  description = "DigitalOcean region"
}

variable "do_token" {
  type      = string
  sensitive = true
}

variable "pvt_key" {
  type        = string
  description = "Private SSH key path"
  default     = "~/.ssh/leadstorefront"
}

variable "pub_key" {
  type        = string
  description = "Public SSH key path"
  default     = "~/.ssh/leadstorefront.pub"
}

variable "docker_username" {
  type        = string
  description = "DockerHub username"
}

variable "docker_password" {
  type        = string
  sensitive   = true
  description = "DockerHub password or token"
}

variable "github_username" {
  type        = string
  description = "GitHub username"
  default     = ""
}

variable "github_password" {
  type        = string
  sensitive   = true
  description = "GitHub password or token"
  default     = ""
}

variable "my_ip" {
  description = "Your current public IP with /32"
  type        = string
}
