resource "yandex_function" "tg_update_handler" {
  name               = "tg-update-handler"
  user_hash          = data.external.function_zip["tg_update_handler"].result["hash"]
  runtime            = "golang121"
  entrypoint         = "handlers/api.HandleTelegramUpdate"
  memory             = "128"
  execution_timeout  = "10"
  service_account_id = yandex_iam_service_account.tgbot_sa.id
  environment = {
    "TG_UPDATES_MQ_ENABLED"   = "true"
    "TG_UPDATES_MQ_ENDPOINT"  = "https://message-queue.api.cloud.yandex.net/"
    "TG_UPDATES_MQ_QUEUE_URL" = yandex_message_queue.telegram_updates_queue.id
  }
  secrets {
    environment_variable = "DB_AWS_KEY_ID"
    id                   = yandex_lockbox_secret.tgbot_sa_static_key_secret.id
    version_id           = yandex_lockbox_secret_version.tgbot_sa_static_key_secret_version.id
    key                  = "ACCESS_KEY_ID"
  }
  secrets {
    environment_variable = "DB_AWS_SECRET_KEY"
    id                   = yandex_lockbox_secret.tgbot_sa_static_key_secret.id
    version_id           = yandex_lockbox_secret_version.tgbot_sa_static_key_secret_version.id
    key                  = "SECRET_ACCESS_KEY"
  }
  content {
    zip_filename = "${path.module}/zips/tg_update_handler.zip"
  }
}

resource "yandex_function" "splitwise_oauth" {
  name               = "splitwise-oauth"
  user_hash          = data.external.function_zip["splitwise_oauth"].result["hash"]
  runtime            = "golang121"
  entrypoint         = "handlers/splitwise_handler.HandleSplitwiseCallback"
  memory             = "128"
  execution_timeout  = "30"
  service_account_id = yandex_iam_service_account.tgbot_sa.id
  environment = {
    "TG_UPDATES_MQ_ENABLED"   = "true"
    "TG_UPDATES_MQ_ENDPOINT"  = "https://message-queue.api.cloud.yandex.net/"
    "TG_UPDATES_MQ_QUEUE_URL" = yandex_message_queue.telegram_updates_queue.id
  }
  secrets {
    environment_variable = "DB_AWS_KEY_ID"
    id                   = yandex_lockbox_secret.tgbot_sa_static_key_secret.id
    version_id           = yandex_lockbox_secret_version.tgbot_sa_static_key_secret_version.id
    key                  = "ACCESS_KEY_ID"
  }
  secrets {
    environment_variable = "DB_AWS_SECRET_KEY"
    id                   = yandex_lockbox_secret.tgbot_sa_static_key_secret.id
    version_id           = yandex_lockbox_secret_version.tgbot_sa_static_key_secret_version.id
    key                  = "SECRET_ACCESS_KEY"
  }
  content {
    zip_filename = "${path.module}/zips/splitwise_oauth.zip"
  }
}

resource "yandex_function" "tg_update_processor" {
  name               = "tg-update-processor"
  user_hash          = data.external.function_zip["tg_update_processor"].result["hash"]
  runtime            = "golang121"
  entrypoint         = "handlers/tg_update_processor.ProcessTelegramUpdate"
  memory             = "128"
  execution_timeout  = "300"
  service_account_id = yandex_iam_service_account.tgbot_sa.id

  environment = {
    "DB_ENDPOINT" = yandex_ydb_database_serverless.ydb_serverless_prod.document_api_endpoint

    "OAI_ENABLED"          = "true"
    "OAI_API_ENDPOINT"     = "https://api.openai.com/v1/"
    "OAI_WHISPER_MODEL_ID" = "whisper-1"

    "OCR_CLIENT" = "gpt"

    "TOKENS_TABLE_NAME" = "tokens"
    "USERS_TABLE_NAME"  = "users"

    "SPLITWISE_ENABLED"      = "true"
    "SPLITWISE_REDIRECT_URL" = "https://splitwiseai.sphericalpotatoinvacuum.xyz/splitwise"

    "WEB_APP_URL" = "https://splitwiseai.sphericalpotatoinvacuum.xyz/webapp"
  }
  secrets {
    environment_variable = "DB_AWS_KEY_ID"
    id                   = yandex_lockbox_secret.tgbot_sa_static_key_secret.id
    version_id           = yandex_lockbox_secret_version.tgbot_sa_static_key_secret_version.id
    key                  = "ACCESS_KEY_ID"
  }
  secrets {
    environment_variable = "DB_AWS_SECRET_KEY"
    id                   = yandex_lockbox_secret.tgbot_sa_static_key_secret.id
    version_id           = yandex_lockbox_secret_version.tgbot_sa_static_key_secret_version.id
    key                  = "SECRET_ACCESS_KEY"
  }
  secrets {
    environment_variable = "MINDEE_API_TOKEN"
    id                   = "e6ql7hebo9gikq0cfgcr"
    version_id           = "e6qnrbsv41nq1h6sied0"
    key                  = "TOKEN"
  }
  secrets {
    environment_variable = "OAI_API_TOKEN"
    id                   = "e6q3a53fhcq93o5nbjkf"
    version_id           = "e6q57irdm1sr9kggd1ip"
    key                  = "KEY"
  }
  secrets {
    environment_variable = "SPLITWISE_API_TOKEN"
    id                   = "e6qeat2rr6muotl3etat"
    version_id           = "e6qckedenmuc6nfegtlh"
    key                  = "TOKEN"
  }
  secrets {
    environment_variable = "SPLITWISE_CLIENT_ID"
    id                   = "e6qeat2rr6muotl3etat"
    version_id           = "e6qckedenmuc6nfegtlh"
    key                  = "CLIENT_ID"
  }
  secrets {
    environment_variable = "SPLITWISE_CLIENT_SECRET"
    id                   = "e6qeat2rr6muotl3etat"
    version_id           = "e6qckedenmuc6nfegtlh"
    key                  = "CLIENT_SECRET"
  }
  secrets {
    environment_variable = "TELEGRAM_BOT_TOKEN"
    id                   = "e6qvt0lf6055fqok3uke"
    version_id           = "e6q3ajahhne8994e659l"
    key                  = "TOKEN"
  }
  content {
    zip_filename = "${path.module}/zips/tg_update_processor.zip"
  }
}

variable "functions" {
  description = "A map of function names to their source files and directories"
  type        = map(list(string))
  default = {
    "tg_update_handler" = [
      "internal",
      "handlers/ext",
      "handlers/api.go",
      "go.mod",
      "go.sum",
    ]
    "splitwise_oauth" = [
      "internal",
      "handlers/ext",
      "handlers/splitwise_handler.go",
      "go.mod",
      "go.sum",
    ]
    "tg_update_processor" = [
      "internal",
      "handlers/ext",
      "handlers/tg_update_processor.go",
      "go.mod",
      "go.sum",
    ]
  }
}

data "external" "function_zip" {
  for_each = var.functions

  program = ["bash", "${path.module}/zip_functions.sh", "deploy/zips/${each.key}.zip", "../", join(",", each.value)]
}
