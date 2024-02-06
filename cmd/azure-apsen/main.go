package main

import (
	"azure-secret-expiration-notifier/internal"
	"flag"
	"github.com/robfig/cron/v3"
	"log"
	"slices"
)

func RunChecks(config *internal.Config, azClient *internal.AzureClient) {
	intervals := append([]int{0}, config.Checks.ThresholdDays...)
	slices.Sort(intervals)

	log.Printf("Checking day intervals %v\n", intervals)

	buckets := internal.NewBuckets(intervals)

	log.Printf("Authenticating azure client using [tenant_id=%s, client_id=%s]\n", azClient.TenantId, azClient.Auth.ClientId)
	if err := azClient.Authenticate(); err != nil {
		log.Fatalf("Failed authenticate Azure client: %s\n", err)
		return
	}

	apps, _ := azClient.ListApplications()
	secrets := internal.FlattenAndSortSecrets(apps)

	log.Printf("Found %d apps with %d secrets\n", len(apps), len(secrets))

	for _, sec := range secrets {
		buckets.Put(sec)
	}

	if config.Notifications == nil || len(config.Notifications.Providers) == 0 {
		return
	}

	log.Println("Sending notifications:")

	notifyConfig := config.Notifications.Providers

	for _, providerConfig := range notifyConfig {
		var provider internal.Notifier

		switch providerConfig.Type {
		case "slack":
			{
				bd := providerConfig.Values

				provider = &internal.SlackNotifier{
					ApiKey:    bd["token"].AsString(),
					ChannelId: bd["channel"].AsString(),
				}
			}

		case "console":
			{
				provider = &internal.ConsoleNotifier{}
			}
		}

		if provider == nil {
			continue
		}
		log.Printf("Sending '%s' notification.\n", providerConfig.Type)

		provider.Notify(buckets)
	}

	log.Println("Done.")
}

func main() {
	configPath := flag.String("config", "config.hcl", "configuration file path.")

	log.Printf("Loading config from '%s'", *configPath)
	config := internal.MustLoadConfig(*configPath)

	az := config.Azure

	azClient := internal.NewAzureClient(az.TenantId,
		internal.SimpleAzureAuth(az.ClientId, az.ClientSecret),
	)

	if config.Notifications == nil || len(config.Notifications.Providers) == 0 {
		log.Println("No notifications will be sent as no providers are configured.")
	}

	c := cron.New()

	log.Printf("Scheduling cron Job: '%s'\n", config.Checks.ScheduleCron)
	_, err := c.AddFunc(config.Checks.ScheduleCron, func() {
		RunChecks(config, azClient)
	})

	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	c.Run()

	select {}
}
