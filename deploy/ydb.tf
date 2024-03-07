resource "yandex_ydb_database_serverless" "ydb_serverless_prod" {
  name = "ydb-serverless-prod"

  deletion_protection = true
}
