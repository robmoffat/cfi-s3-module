variable "file_source" {
  description = "Path to the file to upload"
  type        = string
  default     = null
}

variable "bucket_name" {
  description = "Name of the bucket to store the object in"
  type        = string
}

variable "object_key" {
  description = "The name/path the file will have in the bucket"
  type        = string
}

variable "content_type" {
  description = "The content type of the file (e.g., 'application/json', 'image/jpeg')"
  type        = string
  default     = null
}

variable "tags" {
  description = "A map of tags to assign to the object"
  type        = map(string)
  default     = {}
}

