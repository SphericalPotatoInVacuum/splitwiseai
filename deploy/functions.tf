resource "yandex_function" "tg_bot" {
  name               = "tg-bot"
  user_hash          = data.external.function_zip["tg_bot"].result["hash"]
  runtime            = "golang121"
  entrypoint         = "index.Handler"
  memory             = "128"
  execution_timeout  = "30"
  service_account_id = yandex_iam_service_account.tgbot_sa.id
  environment = {
    "DB_AWS_KEY_ID"          = yandex_iam_service_account_static_access_key.tgbot_sa_static_key_secret.access_key
    "DB_AWS_SECRET_KEY"      = yandex_iam_service_account_static_access_key.tgbot_sa_static_key_secret.secret_key
    "DB_ENDPOINT"            = yandex_ydb_database_serverless.ydb_serverless_prod.document_api_endpoint
    "OAI_API_ENDPOINT"       = "https://api.openai.com/v1/"
    "OAI_WHISPER_MODEL_ID"   = "whisper-1"
    "OCR_CLIENT"             = "gpt"
    "TOKENS_TABLE_NAME"      = "tokens"
    "USERS_TABLE_NAME"       = "users"
    "SPLITWISE_REDIRECT_URL" = "https://splitwiseai.sphericalpotatoinvacuum.xyz/splitwise"
    "WEB_APP_URL"            = "https://splitwiseai.sphericalpotatoinvacuum.xyz/webapp"
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
    zip_filename = "${path.module}/zips/tg_bot.zip"
  }
}

variable "functions" {
  description = "A map of function names to their source files and directories"
  type        = map(list(string))
  default = {
    "tg_bot" = [
      "internal",
      "handlers",
      "go.mod",
      "go.sum",
    ]
  }
}

data "external" "function_zip" {
  for_each = var.functions

  program = ["bash", "${path.module}/zip_functions.sh", "deploy/zips/${each.key}.zip", "../", join(",", each.value)]
}
