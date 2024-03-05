resource "yandex_iam_service_account" "tgbot_sa" {
  name        = "splitwise-tgbot-sa"
  description = "Service account for Splitwise Telegram bot"
}

variable "tgbot_roles" {
  type = set(string)
  default = [
    "ymq.reader",
    "ymq.writer",
    "lockbox.payloadViewer",
    "ydb.editor",
    "ydb.viewer",
  ]
}

resource "yandex_resourcemanager_folder_iam_binding" "tgbot_sa" {
  for_each  = var.tgbot_roles
  role      = each.value
  folder_id = var.folder_id
  members = [
    "serviceAccount:${yandex_iam_service_account.tgbot_sa.id}",
  ]
}

resource "yandex_iam_service_account_static_access_key" "tgbot_sa_static_key_secret" {
  service_account_id = yandex_iam_service_account.tgbot_sa.id
  description        = "Key for YDB access"
}

resource "yandex_lockbox_secret" "tgbot_sa_static_key_secret" {
  name = "serverless secret"
  description = "Static key for YDB access"
}

resource "yandex_lockbox_secret_version" "tgbot_sa_static_key_secret_version" {
  secret_id = yandex_lockbox_secret.tgbot_sa_static_key_secret.id
  entries {
    key        = "ACCESS_KEY_ID"
    text_value = yandex_iam_service_account_static_access_key.tgbot_sa_static_key_secret.access_key
  }
  entries {
    key        = "SECRET_ACCESS_KEY"
    text_value = yandex_iam_service_account_static_access_key.tgbot_sa_static_key_secret.secret_key
  }
}

resource "yandex_iam_service_account" "gateway_api_sa" {
  name        = "splitwise-gateway-api-sa"
  description = "Service account for Gateway API"
}

variable "gateway_api_roles" {
  type = set(string)
  default = [
    "serverless.functions.invoker",
  ]
}

resource "yandex_resourcemanager_folder_iam_binding" "gateway_api_sa" {
  for_each  = var.gateway_api_roles
  role      = each.value
  folder_id = var.folder_id
  members = [
    "serviceAccount:${yandex_iam_service_account.gateway_api_sa.id}",
  ]
}
