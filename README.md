# Application Secret Expiration Notifier for Microsoft Azure

This application will check your Azure Application Secrets/Tokens and send a notification about their expiration date.
The parameter `checks.threshold_days` controls how many days in advance a secret will be reported before it's expiration.

**Build**
    
    $ go build -o aspen cmd/azure-apsen/main.go

**Usage:**
```bash
./aspen --help
Usage of ./aspen:
  -config string
        configuration file path. (default "config.hcl")
```

**Example Configuration:**
```hcl
azure {
  tenant_id     = ""
  client_id     = ""
  client_secret = ""
}

checks {
  # Warning intervals
  # warning will be grouped by today to 2 days, 2 days to 7 days, etc…
  threshold_days = [2,7,14]

  # The schedule is controlled with the cron syntax which can be
  # tested at https://crontab.guru/
  schedule_cron = "* 1 * * *" # Will run daily at 1am
}

notifications {
  provider "console" {}
  
  provider "slack" {
    token = "TOKEN"
    channel = "CHANNEL-ID"
  }
}
```