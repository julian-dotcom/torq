package communications

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
)

func SubscribeSlack(ctx context.Context, db *sqlx.DB) {

	serviceType := core.SlackService

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveCoreServiceState(serviceType)
			return
		default:
		}

		log.Debug().Msg("Loading Slack client.")
		socketClient := socketmode.New(getSlackClient(), socketmode.OptionDebug(log.Debug().Enabled()))

		go processEvents(ctx, socketClient, db)

		err := socketClient.RunContext(ctx)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				cache.SetInactiveCoreServiceState(serviceType)
				return
			}
			log.Error().Err(err).Msgf("Disconnected from Slack")
			cache.SetFailedCoreServiceState(serviceType)
			return
		}
	}
}

func getSlackClient() *slack.Client {
	oauth, botToken := cache.GetSettings().GetSlackCredential()
	return slack.New(oauth, slack.OptionDebug(log.Debug().Enabled()), slack.OptionAppLevelToken(botToken))
}

func processEvents(ctx context.Context, socketClient *socketmode.Client, db *sqlx.DB) {
	log.Debug().Msgf("Initiating socketClient.Events for Slack events")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Slack Subscription cancelled")
			return
		case event := <-socketClient.Events:
			switch event.Type {
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Debug().Msgf("Could not type cast the event to the EventsAPIEvent: %v", event)
					continue
				}
				socketClient.Ack(*event.Request)
				handleEventMessage(db, socketClient, eventsAPIEvent)
			case socketmode.EventTypeSlashCommand:
				// Just like before, type cast to the correct event type, this time a SlashEvent
				command, ok := event.Data.(slack.SlashCommand)
				if !ok {
					log.Debug().Msgf("Could not type cast the message to a SlashCommand: %v", command)
					continue
				}
				socketClient.Ack(*event.Request)
				handleSlashCommand(db, command)
			default:
				log.Trace().Msgf("Could not type cast the event.Type: %v", event.Type)
				log.Trace().Msgf("Could not type cast the event.Data: %v", event.Data)
				log.Trace().Msgf("Could not type cast the event.Request: %v", event.Request)
			}
		}
	}
}

func SendSlackBotMessages(botMessage MessageForBot) {
	log.Debug().Msgf("Sending out slack message to %v: %v", botMessage.Slack.Channel, botMessage.Message)
	attachment := slack.Attachment{
		Text: botMessage.Message,
	}
	if botMessage.Slack.Pretext != "" {
		attachment.Pretext = botMessage.Slack.Pretext
	}
	if botMessage.Slack.Color != "" {
		attachment.Color = botMessage.Slack.Color
	}
	_, _, err := getSlackClient().PostMessage(botMessage.Slack.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Error().Err(err).Msgf("Slack bot Send failed: %v", botMessage.Message)
	}
}

func handleSlashCommand(db *sqlx.DB, command slack.SlashCommand) {
	messageForBot := MessageForBot{
		Slack: MessageForSlack{
			Channel: command.ChannelID,
			ReplyTo: command.UserName,
			Color:   "#283B4C",
		},
	}
	HandleMessage(db, messageForBot, command.Text, command.Command[1:], CommunicationSlack)
}

func handleEventMessage(db *sqlx.DB, socketClient *socketmode.Client, event slackevents.EventsAPIEvent) {
	if event.Type != slackevents.CallbackEvent {
		return
	}

	innerEvent := event.InnerEvent
	ev, ok := innerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok {
		return
	}

	user, err := socketClient.GetUserInfo(ev.User)
	if err != nil {
		log.Error().Err(err).Msgf("Slack bot GetUserInfo failed: %v", ev.User)
	}
	messageForBot := MessageForBot{
		Slack: MessageForSlack{
			Channel: ev.Channel,
			ReplyTo: user.Profile.DisplayName,
			Color:   "#4af030",
		},
	}
	eventText := strings.TrimSpace(ev.Text)
	command := extractCommand(eventText)
	if command != "" && strings.LastIndex(eventText, command) != -1 {
		HandleMessage(db, messageForBot,
			strings.TrimSpace(eventText[strings.LastIndex(eventText, command)+len(command):]),
			command, CommunicationSlack)
	}
}

func extractCommand(eventText string) string {
	for _, button := range getButtons() {
		if strings.Contains(eventText, "/"+button) || strings.Contains(eventText, " "+button) {
			return button
		}
	}
	return ""
}
