package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/noxer/eventbot/discord"
)

const timeFormat = "15:04"

var months = []string{
	"Januar",
	"Februar",
	"März",
	"April",
	"Mai",
	"Juni",
	"Juli",
	"August",
	"September",
	"Oktober",
	"November",
	"Dezember",
}

func init() {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		fmt.Printf("Error loading timezone data: %s\n", err)
		return
	}

	time.Local = loc
}

type StateStore map[string]string

func (s StateStore) Get(key string) string {
	return s[key]
}

func (s StateStore) Set(key, value string) {
	old := s[key]
	if old == value {
		return
	}

	s[key] = value
	p, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		fmt.Printf("Error marshaling state: %s\n", err)
		return
	}

	err = os.WriteFile("state.json", p, 0644)
	if err != nil {
		fmt.Printf("Error writing state: %s\n", err)
	}
}

func main() {
	discordSecret := os.Getenv("DISCORD_TOKEN")
	discordGuildID := "278161377045118977"
	discordChannelID := "774183350499803178"

	d, err := discord.New(discordSecret, make(StateStore))
	if err != nil {
		fmt.Printf("Error creating discord client: %s\n", err)
		os.Exit(1)
	}

	for range time.Tick(60 * time.Second) {
		es, err := d.EventsList(discordGuildID)
		if err != nil {
			fmt.Printf("Error retrieving events list: %s\n", err)
			os.Exit(1)
		}

		sb := &strings.Builder{}
		sb.WriteString("**Eventplan:**\n")

		date := ""

		if len(es) == 0 {
			sb.WriteString("Keine Events")
		}

		for _, e := range es {
			if e.Status != discord.StatusScheduled && e.Status != discord.StatusActive {
				continue
			}

			start := e.Start.Local()

			currentStartDate := formatDate(start)
			if date != currentStartDate {
				date = currentStartDate
				sb.WriteString(fmt.Sprintf("\n**%s:**\n", date))
			}

			link := ""
			if e.Type == discord.TypeExternal {
				link = " (=> <" + e.Link + ">)"
			}

			fmt.Fprintf(sb, "***%s Uhr:*** %s%s\n", start.Format(timeFormat), e.Name, link)
			if e.Description != "" {
				sb.WriteByte('_')
				sb.WriteString(strings.TrimSpace(e.Description))
				sb.WriteString("_\n")
			}
		}

		err = d.SendOrUpdateMessage(discordChannelID, sb.String())
		if err != nil {
			fmt.Printf("Error sending message: %s\n", err)
			os.Exit(1)
		}
	}
}

func formatDate(t time.Time) string {
	return fmt.Sprintf("%02d. %s", t.Day(), months[t.Month()-1])
}
