resource "yandex_message_queue" "telegram_updates_queue" {
  name                       = "telegram-updates-queue"
  visibility_timeout_seconds = 600
  receive_wait_time_seconds  = 10
  message_retention_seconds  = 1209600
  access_key                 = yandex_iam_service_account_static_access_key.ymq_trigger_sa_static_key.access_key
  secret_key                 = yandex_iam_service_account_static_access_key.ymq_trigger_sa_static_key.secret_key
}
