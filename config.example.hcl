azure {
  tenant_id     = ""
  client_id     = ""
  client_secret = ""
}

checks {
  # Warning intervals
  threshold_days = [2,7,14]

  # Will run daily at 1am
  schedule_cron = "* 1 * * *"
}

notifications {
  provider "slack" {
    token = "TOKEN"
    channel = "CHANNEL-ID"
  }
}