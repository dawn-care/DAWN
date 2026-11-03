variable "region" {
  description = "AWS region for this facility."
  type        = string
  default     = "us-east-1"
}

variable "facility" {
  description = "Short facility identifier, for example stanns."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{0,30}$", var.facility))
    error_message = "Use lowercase letters, numbers and hyphens, starting with a letter (maximum 31 characters)."
  }
}

variable "ami_id" {
  description = "Pinned official Canonical Ubuntu Server 24.04 LTS x86_64 AMI in the selected region."
  type        = string

  validation {
    condition     = can(regex("^ami-([0-9a-f]{8}|[0-9a-f]{17})$", var.ami_id))
    error_message = "Replace the placeholder with an actual Ubuntu AMI ID."
  }
}

variable "instance_type" {
  description = "EC2 size; t3.medium provides 2 vCPU and 4 GiB RAM."
  type        = string
  default     = "t3.medium"
}

variable "storage_gb" {
  description = "Encrypted root disk, including Docker images and persistent volumes."
  type        = number
  default     = 50

  validation {
    condition     = var.storage_gb >= 50 && floor(var.storage_gb) == var.storage_gb
    error_message = "Storage must be a whole number of GiB, at least 50."
  }
}

variable "admin_cidr" {
  description = "Administrator's public IPv4 address with /32; only this address can SSH."
  type        = string

  validation {
    condition     = can(cidrnetmask(var.admin_cidr)) && endswith(var.admin_cidr, "/32")
    error_message = "Supply one public IPv4 address with /32, not an unrestricted SSH range."
  }
}

variable "ssh_public_key" {
  description = "Administrator's SSH public key contents. Never supply the private key."
  type        = string

  validation {
    condition     = can(regex("^(ssh-ed25519|ssh-rsa|ecdsa-sha2-nistp256) [A-Za-z0-9+/=]+", var.ssh_public_key))
    error_message = "Replace the placeholder with an SSH public key."
  }
}

variable "dns_zone_id" {
  description = "Existing shared Route 53 public hosted zone ID; null skips DNS creation."
  type        = string
  default     = null
}

variable "hostname" {
  description = "Facility hostname, for example stanns.dawn-care.com; required when dns_zone_id is set."
  type        = string
  default     = null
}
