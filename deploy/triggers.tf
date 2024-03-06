resource "yandex_function_trigger" "ymq_trigger" {
  name        = "ymq-trigger"
  description = "Trigger to process telegram updates from YMQ"
  function {
    id                 = yandex_function.tg_update_processor.id
    tag                = "$latest"
    service_account_id = yandex_iam_service_account.ymq_trigger_sa.id
  }
  message_queue {
    queue_id           = yandex_message_queue.telegram_updates_queue.arn
    service_account_id = yandex_iam_service_account.ymq_trigger_sa.id
    batch_cutoff       = 1
    batch_size         = 1
  }
}
