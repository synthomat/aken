package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Notifier interface {
	Notify(buckets *Buckets)
}

type SlackNotifier struct {
	ApiKey    string
	ChannelId string
}

type js map[string]any

func Buckets2SlackMsg(buckets *Buckets) []any {
	message := []any{
		js{
			"type": "section",
			"text": js{
				"type": "mrkdwn",
				"text": "🗓️ Here comes the *App Secret Expiration Report*",
			},
		},
	}

	for _, bucket := range *buckets {
		if len(bucket.Secrets) == 0 {
			continue
		}

		message = append(message, js{
			"type": "header",
			"text": js{
				"type": "plain_text",
				"text": fmt.Sprintf("%s – %s", bucket.From.Format("2006-01-02"), bucket.To.Format("2006-01-02")),
			},
		})

		for _, sec := range bucket.Secrets {
			message = append(message, js{
				"type": "section",
				"text": js{
					"type": "mrkdwn",
					"text": fmt.Sprintf("<https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationMenuBlade/~/Credentials/appId/%s|*%s*> / %s: %s",
						sec.App.AppId, sec.App.DisplayName, sec.DisplayName, sec.EndDateTime.Format("2006-01-02 15:04:05")),
				},
			})
		}
	}

	message = append(message, js{
		"type": "divider",
	})

	return message
}

func (sn *SlackNotifier) Notify(buckets *Buckets) {

	slackMessage := js{
		"channel": sn.ChannelId,
		"blocks":  Buckets2SlackMsg(buckets),
	}

	plainJson, _ := json.Marshal(slackMessage)

	req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(plainJson))
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sn.ApiKey)
	c := http.Client{}

	resp, err := c.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var parsedBody map[string]any

	json.Unmarshal(body, &parsedBody)

	if parsedBody["ok"] == false {
		log.Printf("Failed sending Slack notification: %s\n", string(body))
	}
}

type ConsoleNotifier struct {
}

func (cn *ConsoleNotifier) Notify(buckets *Buckets) {
	log.Println("– Start Report –")

	for _, b := range *buckets {
		log.Printf("+ %s – %s\n", b.From.Format("2006-01-02"), b.To.Format("2006-01-02"))

		for _, s := range b.Secrets {
			log.Printf("|  %s / %s: %s\n", s.App.DisplayName, s.DisplayName, s.EndDateTime.Format("2006-01-02 15:04:05"))
		}
		log.Println("|")
	}
	log.Println("– End Report –")
}
