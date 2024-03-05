resource "yandex_api_gateway" "splitwise_api_gateway" {
  name = "splitwise-api-gateway"
  spec = <<-EOT
    openapi: "3.0.0"
    info:
      version: 1.0.0
      title: Splitwise AI API
    paths:
      /update:
        post:
          summary: Path for telegram updates
          operationId: telegram-update
          x-yc-apigateway-integration:
            type: cloud_functions
            function_id: ${yandex_function.tg_bot.id}
            service_account_id: ${yandex_iam_service_account.gateway_api_sa.id}
      /splitwise:
        get:
          summary: Path for splitwise requests
          operationId: splitwise-request
          x-yc-apigateway-integration:
            type: cloud_functions
            function_id: ${yandex_function.tg_bot.id}
            service_account_id: ${yandex_iam_service_account.gateway_api_sa.id}
  EOT
  custom_domains {
    certificate_id = "fpqdr0cund0je26ttp4q"
    domain_id      = "d5dpknrjngekk8qf33q1"
    fqdn           = "splitwiseai.sphericalpotatoinvacuum.xyz"
  }
}
