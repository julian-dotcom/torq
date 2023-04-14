package communications

import (
	"context"
	"fmt"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
)

// parseMode == ModeHTML
//
//	"<", "&lt;", ">", "&gt;", "&", "&amp;"
//
//	parseMode == ModeMarkdown {
//	 "_", "\\_", "*", "\\*", "`", "\\`", "[", "\\["
//
// parseMode == ModeMarkdownV2
//
//	"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(",
//	"\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>",
//	"#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
//	"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",

const (
	// Button texts
	statusText   = "status ✅"
	registerText = "register ⚡️"
	settingsText = "settings ⚙️"

	deactivateNodeDetailText = "Node details 🛑"

	activateNodeDetailText = "Node details 🟢"

	SupportLink = "https://t.me/joinchat/V-Dks6zjBK4xZWY0"
	SupportText = "LN.capital telegram channel"
)

func getMainMenu() tgbotapi.InlineKeyboardMarkup {
	// Keyboard layout for the first menu.
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(registerText, RegisterButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(statusText, StatusButton),
			tgbotapi.NewInlineKeyboardButtonData(settingsText, SettingsButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(SupportText, SupportLink),
		),
	)
}

func getSupportMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(SupportText, SupportLink),
		),
	)
}

func getMenus() map[Menu]tgbotapi.InlineKeyboardMarkup {
	return map[Menu]tgbotapi.InlineKeyboardMarkup{
		MenuMain:    getMainMenu(),
		MenuSupport: getSupportMenu(),
	}
}

func getMenuHeadings() map[Menu]string {
	return map[Menu]string{
		MenuMain:    "Hello!\nType /help to if you need help.",
		MenuSupport: "\nWhen in doubt contact us in our telegram channel",
	}
}

var (
	telegramHighPriorityOnce   sync.Once //nolint:gochecknoglobals
	telegramHighPriorityObject *Telegram //nolint:gochecknoglobals
	telegramLowPriorityOnce    sync.Once //nolint:gochecknoglobals
	telegramLowPriorityObject  *Telegram //nolint:gochecknoglobals
)

type Telegram struct {
	bot *tgbotapi.BotAPI
}

func getTelegramHighPriority() (*Telegram, error) {
	telegramHighPriorityOnce.Do(func() {
		log.Debug().Msg("Loading TelegramHighPriority client.")
		bot, err := tgbotapi.NewBotAPI(cache.GetSettings().GetTelegramCredential(true))
		telegramHighPriorityObject = &Telegram{
			bot: bot,
		}
		if err != nil {
			log.Error().Err(err).Msgf("Failed to initialize TelegramHighPriority bot.")
		}
	})
	return telegramHighPriorityObject, nil
}

func getTelegramLowPriority() (*Telegram, error) {
	telegramLowPriorityOnce.Do(func() {
		log.Debug().Msg("Loading TelegramLowPriority client.")
		bot, err := tgbotapi.NewBotAPI(cache.GetSettings().GetTelegramCredential(false))
		bot.Debug = log.Debug().Enabled()
		telegramLowPriorityObject = &Telegram{
			bot: bot,
		}
		if err != nil {
			log.Error().Err(err).Msgf("Failed to initialize TelegramLowPriority bot.")
		}
	})
	return telegramLowPriorityObject, nil
}

func SubscribeTelegram(ctx context.Context, db *sqlx.DB, highPriority bool) {
	serviceType := core.TelegramHighService
	communicationTargetType := CommunicationTelegramHighPriority
	if !highPriority {
		serviceType = core.TelegramLowService
		communicationTargetType = CommunicationTelegramLowPriority
	}

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveCoreServiceState(serviceType)
			return
		default:
		}

		telegram, err := getTelegramBot(communicationTargetType)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to obtain telegram bot")
			cache.SetFailedCoreServiceState(serviceType)
			return
		}

		log.Debug().Msgf("Initiating tgbotapi.NewUpdate for Telegram events (highPriority: %v)", highPriority)
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = 30
		updates := telegram.bot.GetUpdatesChan(updateConfig)
		for {
			select {
			case <-ctx.Done():
				log.Info().Msgf("Telegram Subscription cancelled (SubscribeTelegram highPriority: %v)", highPriority)
				cache.SetInactiveCoreServiceState(serviceType)
				return
			// receive update from channel and then handle it
			case update := <-updates:
				handleUpdate(db, update, communicationTargetType)
			}
		}
	}
}

func getTelegramBot(targetType CommunicationTargetType) (*Telegram, error) {
	switch targetType {
	case CommunicationTelegramHighPriority:
		return getTelegramHighPriority()
	case CommunicationTelegramLowPriority:
		return getTelegramLowPriority()
	}
	return nil, nil
}

func SendTelegramBotMessages(botMessage MessageForBot, targetType CommunicationTargetType) {
	telegram, err := getTelegramBot(targetType)
	if err != nil {
		log.Error().Err(err).Msgf("Telegram bot connection failed")
		return
	}
	if botMessage.Telegram.ParseMode == "" || botMessage.Telegram.ParseMode == tgbotapi.ModeMarkdownV2 {
		escapes := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
		for _, e := range escapes {
			botMessage.Message = strings.ReplaceAll(botMessage.Message, e, "\\"+e)
		}
	}
	log.Info().Msgf("Sending out telegram message to %v: %v", botMessage.Telegram.Id, botMessage.Message)
	if botMessage.HasMessage() {
		msg := tgbotapi.NewMessage(botMessage.Telegram.Id, "")
		msg.Text = botMessage.Message
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		if botMessage.Telegram.ParseMode != "" {
			msg.ParseMode = botMessage.Telegram.ParseMode
		}
		if botMessage.Telegram.ReplyToMessageId != 0 {
			msg.ReplyToMessageID = botMessage.Telegram.ReplyToMessageId
		}
		if botMessage.Telegram.ReplyMarkup != nil {
			msg.ReplyMarkup = *botMessage.Telegram.ReplyMarkup
		}
		_, err = telegram.bot.Send(msg)
		if err != nil {
			log.Error().Err(err).Msgf("Telegram bot Send failed: %v", msg.Text)
		}
	}
	if botMessage.HasMenu() {
		for _, menu := range botMessage.Menus {
			telegramMenu, exists := getMenus()[menu]
			if exists {
				menuMsg := tgbotapi.NewMessage(botMessage.Telegram.Id, "")
				menuMsg.ParseMode = tgbotapi.ModeHTML
				if botMessage.Telegram.ReplyToMessageId != 0 {
					menuMsg.ReplyToMessageID = botMessage.Telegram.ReplyToMessageId
				}
				menuMsg.Text = botMessage.Message + getMenuHeadings()[menu]
				menuMsg.ReplyMarkup = telegramMenu
				_, err = telegram.bot.Send(menuMsg)
				if err != nil {
					log.Error().Err(err).Msgf("Telegram bot Send failed: %v", menuMsg.Text)
				}
			}
		}
	}
	if botMessage.Error != "" {
		menuMsg := tgbotapi.NewMessage(botMessage.Telegram.Id, "")
		menuMsg.ParseMode = tgbotapi.ModeHTML
		if botMessage.Telegram.ReplyToMessageId != 0 {
			menuMsg.ReplyToMessageID = botMessage.Telegram.ReplyToMessageId
		}
		menuMsg.ReplyMarkup = getSupportMenu()
		_, err = telegram.bot.Send(menuMsg)
		if err != nil {
			log.Error().Err(err).Msgf("Telegram bot Send failed: %v", menuMsg.Text)
		}
	}
}

func handleUpdate(db *sqlx.DB, update tgbotapi.Update, communicationTargetType CommunicationTargetType) {
	switch {
	case update.Message != nil:
		messageForBot := MessageForBot{
			Telegram: MessageForTelegram{
				Id:       update.Message.Chat.ID,
				UserName: update.Message.Chat.UserName,
			},
		}
		var command string
		text := strings.TrimSpace(update.Message.Text)
		for _, button := range getButtons() {
			if strings.Contains(text, "/"+button) || strings.Contains(text, " "+button) {
				command = button
				index := strings.LastIndex(text, button) + len(button)
				if index < len(text) {
					text = text[index:]
				} else {
					text = ""
				}
				text = strings.TrimSpace(text)
				HandleMessage(db, messageForBot, text, command, communicationTargetType)
				break
			}
		}
	case update.CallbackQuery != nil:
		messageForBot := MessageForBot{
			Telegram: MessageForTelegram{
				Id:               update.CallbackQuery.Message.Chat.ID,
				UserName:         update.CallbackQuery.Message.Chat.UserName,
				ReplyToMessageId: update.CallbackQuery.Message.MessageID,
			},
		}
		var command string
		text := update.CallbackQuery.Data
		for _, button := range getButtons() {
			if strings.Contains(text, "/"+button) || strings.Contains(text, " "+button) {
				command = button
				index := strings.LastIndex(text, button) + len(button)
				if index < len(text) {
					text = text[index:]
				} else {
					text = ""
				}
				text = strings.TrimSpace(text)
				HandleButton(db, messageForBot, command, text, communicationTargetType)
				break
			}
		}
	case update.ChannelPost != nil:
		messageForBot := MessageForBot{
			Telegram: MessageForTelegram{
				Id:       update.ChannelPost.Chat.ID,
				UserName: update.ChannelPost.Chat.Title,
			},
		}
		HandleMessage(db, messageForBot, update.ChannelPost.Text, update.ChannelPost.Command(), communicationTargetType)
	}
}

func SendNodeSettingsMenu(db *sqlx.DB,
	communicationTargetType CommunicationTargetType,
	messageForBot MessageForBot) {

	publicKeyMsg := "Node Settings (all channels for this node)\n🟢 = Activate, 🛑 = Deactivate\n" +
		"The public key is required when multiple nodes are registered.\n" +
		" To set up the public key, use the command /publickey $PUBLICKEY\n\n"
	var err error
	var communicationId int
	publicKey := PublicKeys[communicationTargetType][messageForBot.GetChannelIdentifier()]
	if publicKey != "" {
		specifiedNodeId := cache.GetChannelPeerNodeIdByPublicKey(publicKey, core.Bitcoin, core.MainNet)
		nodeIds, err := GetNodeIdsByCommunication(db, communicationTargetType)
		if err != nil {
			log.Error().Err(err).Msg("Telegram bot failed to obtain existing nodeId")
		}
		var nodeId int
		if slices.Contains(nodeIds, specifiedNodeId) {
			nodeId = specifiedNodeId
		}
		if nodeId == 0 {
			messageForBot.Message = fmt.Sprintf("Public key: %v is not registered", publicKey)
			SendTelegramBotMessages(messageForBot, communicationTargetType)
			return
		}
		communicationIds, err := GetCommunicationIdsByNodeId(db, nodeId, communicationTargetType)
		if err != nil {
			log.Error().Err(err).Msg("Telegram bot failed to obtain existing nodeId")
		}
		if len(communicationIds) != 0 {
			communicationId = communicationIds[0]
		}
	}
	communicationIds, err := GetCommunicationIdsByCommunicationTargetType(db, communicationTargetType)
	if err != nil {
		log.Error().Err(err).Msg("Telegram bot failed to obtain existing nodeId")
	}
	if len(communicationIds) != 0 {
		communicationId = communicationIds[0]
	}

	if communicationId == 0 {
		messageForBot.Message = "/register > Node Registration"
		SendTelegramBotMessages(messageForBot, communicationTargetType)
		return
	}
	settings, err := GetCommunicationSettings(db, communicationId)
	if err != nil {
		log.Error().Err(err).Msg("Telegram bot failed to obtain existing settings")
	}
	messageForBot.Message = publicKeyMsg
	markup := getNodeSettingsMenuMarkup(settings[NodeDetailsChanged])
	messageForBot.Telegram.ReplyMarkup = &markup
	messageForBot.Telegram.ParseMode = tgbotapi.ModeHTML
	SendTelegramBotMessages(messageForBot, communicationTargetType)
}

func getNodeSettingsMenuMarkup(nodeDetailsActive bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var button tgbotapi.InlineKeyboardButton
	if nodeDetailsActive {
		button = tgbotapi.NewInlineKeyboardButtonData(deactivateNodeDetailText, DeactivateNodeDetailButton)
	} else {
		button = tgbotapi.NewInlineKeyboardButtonData(activateNodeDetailText, ActivateNodeDetailButton)
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
