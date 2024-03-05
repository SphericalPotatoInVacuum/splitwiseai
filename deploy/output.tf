output "ydb_docapi_endpoint" {
  value = yandex_ydb_database_serverless.ydb_serverless_prod.document_api_endpoint
}

output "api_gateway_endpoint" {
  value = yandex_api_gateway.splitwise_api_gateway.domain
}
