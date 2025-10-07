variable "project_id" {
  type = string
}

module "sql_database" {
  source  = "git::https://github.com/terraform-google-modules/terraform-google-sql-db.git//examples/postgresql-psc?ref=v26.2.1"
  project_id = var.project_id
}
