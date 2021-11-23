package discord

import (
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

type StateStore interface {
	Set(key, value string)
	Get(key string) string
}

type EventType int

const (
	TypeVoice    EventType = discordgo.GuildScheduledEventEntityTypeVoice
	TypeStaged             = discordgo.GuildScheduledEventEntityTypeStageInstance
	TypeExternal           = discordgo.GuildScheduledEventEntityTypeExternal
)

type EventStatus int

const (
	StatusActive    EventStatus = discordgo.GuildScheduledEventStatusActive
	StatusCompleted             = discordgo.GuildScheduledEventStatusCompleted
	StatusCanceled              = discordgo.GuildScheduledEventStatusCanceled
	StatusScheduled             = discordgo.GuildScheduledEventStatusScheduled
)

type Event struct {
	ID          string
	Name        string
	Status      EventStatus
	Type        EventType
	Description string
	Start       time.Time
	End         time.Time
	Link        string
}

type Client struct {
	session *discordgo.Session
	state   StateStore
}

func New(token string, state StateStore) (*Client, error) {
	d, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Client{
		session: d,
		state:   state,
	}, nil
}

func (c *Client) EventsList(guildID string) ([]Event, error) {
	events, err := c.session.GuildScheduledEvents(guildID)
	if err != nil {
		return nil, err
	}

	sort.Slice(events, func(i, j int) bool {
		istart, _ := events[i].ScheduledStartTime.Parse()
		jstart, _ := events[j].ScheduledStartTime.Parse()

		return istart.Before(jstart)
	})

	eventsList := make([]Event, len(events))
	for i, event := range events {
		eventsList[i] = convertEvent(event)
	}

	return eventsList, nil
}

func convertEvent(e *discordgo.GuildScheduledEvent) Event {
	start, _ := e.ScheduledStartTime.Parse()
	end, _ := e.ScheduledEndTime.Parse()

	return Event{
		ID:          e.ID,
		Name:        e.Name,
		Status:      EventStatus(e.Status),
		Type:        EventType(e.EntityType),
		Description: e.Description,
		Start:       start,
		End:         end,
		Link:        e.EntityMetadata.Location,
	}
}

func (c *Client) SendOrUpdateMessage(channelID, content string) error {
	var (
		msg *discordgo.Message
		err error
	)

	messageBody := c.state.Get("messageBody")
	if messageBody == content {
		return nil
	} else {
		c.state.Set("messageBody", content)
	}

	messageID := c.state.Get("messageID")
	if messageID == "" {
		msg, err = c.session.ChannelMessageSend(channelID, content)
	} else {
		err = c.session.ChannelMessageDelete(channelID, messageID)
		if err != nil {
			return err
		}

		msg, err = c.session.ChannelMessageSend(channelID, content)
	}

	if err != nil {
		return err
	}

	if messageID != msg.ID {
		c.state.Set("messageID", msg.ID)
	}

	return err
}
