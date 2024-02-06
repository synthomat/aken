package internal

import (
	"bytes"
	"encoding/json"
	"github.com/flosch/pongo2/v6"
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

func (sn *SlackNotifier) Notify(buckets *Buckets) {
	tplFile, err := res.ReadFile("resources/slack-notification.json.j2")
	t, _ := pongo2.FromBytes(tplFile)

	slackMessage, _ := t.Execute(pongo2.Context{
		"channelId": sn.ChannelId,
		"buckets":   buckets,
	})
	req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader([]byte(slackMessage)))
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
