package communications

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lightning"
)

type MessageForBot struct {
	Message  string
	Menus    []Menu
	Error    string
	Slack    MessageForSlack
	Telegram MessageForTelegram
}

func (mfb MessageForBot) IsSlack() bool {
	return mfb.Slack.Channel != ""
}

func (mfb MessageForBot) IsTelegram() bool {
	return mfb.Telegram.Id != 0
}

func (mfb MessageForBot) HasMessage() bool {
	return strings.TrimSpace(mfb.Message) != ""
}

func (mfb MessageForBot) HasMenu() bool {
	return len(mfb.Menus) > 0
}

func (mfb MessageForBot) GetChannelIdentifier() string {
	if mfb.IsSlack() {
		return mfb.Slack.Channel
	}
	return fmt.Sprintf("%v", mfb.Telegram.Id)
}

type CommunicationTargetType int

const (
	CommunicationTelegramHighPriority = CommunicationTargetType(iota)
	CommunicationTelegramLowPriority
	CommunicationSlack
)

type CommunicationType byte

// When adding here also add to GetCommunicationTypes
const (
	NodeDetailsChanged CommunicationType = 1 << iota
	//ChannelStatusChanged
	//ChannelHtlcChanged
	//ChannelFeeChanged
	//ChannelTimeLockDeltaChanged
	//ChannelFeeMismatch
)

func GetCommunicationType(communicationTypeString string) *CommunicationType {
	var communicationType CommunicationType
	switch communicationTypeString {
	case "nodeDetailsDeactivate", "nodeDetailsActivate":
		communicationType = NodeDetailsChanged
	//case "nodeStatusActivate", "nodeStatusDeactivate", "channelStatusDeactivate", "channelStatusActivate":
	//	communicationType = ChannelStatusChanged
	//case "nodeHtlcDeactivate", "nodeHtlcActivate", "channelHtlcDeactivate", "channelHtlcActivate":
	//	communicationType = ChannelHtlcChanged
	//case "nodeFeeDeactivate", "nodeFeeActivate", "channelFeeDeactivate", "channelFeeActivate":
	//	communicationType = ChannelFeeChanged
	//case "nodeTimeLockDeltaDeactivate", "nodeTimeLockDeltaActivate", "channelTimeLockDeltaDeactivate", "channelTimeLockDeltaActivate":
	//	communicationType = ChannelTimeLockDeltaChanged
	//case "nodeFeeMisMatchDeactivate", "nodeFeeMisMatchActivate", "channelFeeMatchDeactivate", "channelFeeMatchActivate":
	//	communicationType = ChannelFeeMismatch
	default:
		return nil
	}
	return &communicationType
}

func GetCommunicationTypes() []CommunicationType {
	return []CommunicationType{
		NodeDetailsChanged,
	}
}

type MessageForSlack struct {
	Channel string
	ReplyTo string
	Pretext string
	Color   string
}

type MessageForTelegram struct {
	Id               int64
	UserName         string
	ReplyToMessageId int
	ParseMode        string
	ReplyMarkup      *tgbotapi.InlineKeyboardMarkup
}

var PublicKeys = map[CommunicationTargetType]map[string]string{ //nolint:gochecknoglobals
	CommunicationTelegramHighPriority: make(map[string]string),
	CommunicationTelegramLowPriority:  make(map[string]string),
	CommunicationSlack:                make(map[string]string),
}

type Menu string

const (
	MenuMain         Menu = "main"
	MenuSupport      Menu = "support"
	MenuNodeSettings Menu = "nodeSettings"
)

const (
	MenuButton   = "menu"
	VectorButton = "vector"
	StatusButton = "status"
	//PingButton      = "ping"
	RegisterButton   = "register"
	UnregisterButton = "unregister"
	SettingsButton   = "settings"
	PublicKeyButton  = "publickey"

	ActivateNodeDetailButton   = "nodeDetailsActivate"
	DeactivateNodeDetailButton = "nodeDetailsDeactivate"
)

func getButtons() [7]string {
	return [7]string{MenuButton, VectorButton, StatusButton, RegisterButton, UnregisterButton, SettingsButton, PublicKeyButton} //, PingButton}
}

func Notify(ctx context.Context, db *sqlx.DB) {

	serviceType := commons.NotifierService

	informationResponses := make(map[int]commons.InformationResponse)
	graphInSyncTime := make(map[int]time.Time)
	chainInSyncTime := make(map[int]time.Time)

	ticker := clock.New().Tick(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			commons.SetInactiveTorqServiceState(serviceType)
			return
		case <-ticker:
			for _, torqNodeSettings := range commons.GetActiveTorqNodeSettings() {
				communications, err := GetCommunicationsForNodeDetails(db,
					torqNodeSettings.NodeId,
					CommunicationTelegramHighPriority, CommunicationTelegramLowPriority, CommunicationSlack)
				if err != nil {
					log.Error().Err(err).Msgf("Getting communications failed for nodeId: %v",
						torqNodeSettings.NodeId)
					commons.SetFailedTorqServiceState(serviceType)
					return
				}
				if len(communications) == 0 {
					log.Debug().Msgf("Notifier could not find communication settings for nodeId: %v",
						torqNodeSettings.NodeId)
					continue
				}
				var newInformation commons.InformationResponse
				if commons.IsLndServiceActive(torqNodeSettings.NodeId) {
					newInformation, err = lightning.GetInformationRequest(torqNodeSettings.NodeId)
				} else {
					_, exists := informationResponses[torqNodeSettings.NodeId]
					if !exists {
						continue
					}
					err = errors.New("LndService is offline")
					delete(informationResponses, torqNodeSettings.NodeId)
				}
				if err != nil {
					log.Error().Err(err).Msgf(
						"Failed to obtain node information from: %v or publicKey: %v",
						torqNodeSettings.NodeId, torqNodeSettings.PublicKey)
					message := fmt.Sprintf("Could not connect to LND (%v)", torqNodeSettings.PublicKey)
					if torqNodeSettings.Name != nil && *torqNodeSettings.Name != "" {
						message = fmt.Sprintf("Could not connect to LND (%v)", *torqNodeSettings.Name)
					}
					sendBotMessages(message, communications)
					continue
				}
				previousInformation, exists := informationResponses[torqNodeSettings.NodeId]
				if !exists {
					informationResponses[torqNodeSettings.NodeId] = newInformation
					message := fmt.Sprintf("Connected to LND (%v)", torqNodeSettings.PublicKey)
					if torqNodeSettings.Name != nil && *torqNodeSettings.Name != "" {
						message = fmt.Sprintf("Connected to LND (%v)", *torqNodeSettings.Name)
					}
					sendBotMessages(message, communications)
					continue
				}
				var message string
				// TODO FIXME Language from user for translations
				message = compareGraphSyncTime(previousInformation, newInformation, graphInSyncTime, torqNodeSettings, message)
				message = compareChainSyncTime(previousInformation, newInformation, chainInSyncTime, torqNodeSettings, message)
				message = comparePendingChannelCount(previousInformation, newInformation, message)
				message = compareInactiveChannelCount(previousInformation, newInformation, message)
				message = compareActiveChannelCount(previousInformation, newInformation, message)
				message = compareVersion(previousInformation, newInformation, message)
				informationResponses[torqNodeSettings.NodeId] = newInformation
				if message != "" {
					sendBotMessages(message, communications)
				}
			}
		}
	}
}

func HandleNotification(db *sqlx.DB, notifierEvent commons.NotifierEvent) {
	var err error
	var communications []Communication
	switch notifierEvent.NotificationType {
	case commons.NodeDetails:
		communications, err = GetCommunicationsForNodeDetails(db,
			notifierEvent.NodeId,
			CommunicationTelegramHighPriority, CommunicationTelegramLowPriority, CommunicationSlack)
	}
	if err != nil {
		log.Error().Err(err).Msgf(
			"Getting user communications for Telegram (high) nodeId: %v", notifierEvent.NodeId)
		return
	}
	if len(communications) == 0 {
		log.Debug().Msgf("Notifier could not find communication settings for %v", notifierEvent)
		return
	}
	if notifierEvent.Notification != nil && *notifierEvent.Notification != "" {
		sendBotMessages(*notifierEvent.Notification, communications)
		return
	}
	if notifierEvent.NodeGraphEvent != nil && (*notifierEvent.NodeGraphEvent).NodeId != 0 {
		nodeGraphEvent := notifierEvent.NodeGraphEvent
		var message string
		// TODO FIXME Language from user for translations
		if strings.TrimSpace(nodeGraphEvent.Color) != strings.TrimSpace(nodeGraphEvent.PreviousEventData.Color) {
			message = message + fmt.Sprintf("Color changed from %v to %v\n",
				strings.Replace(nodeGraphEvent.Color, "#", "", -1),
				strings.Replace(nodeGraphEvent.PreviousEventData.Color, "#", "", -1))
		}
		if strings.TrimSpace(nodeGraphEvent.Alias) != strings.TrimSpace(nodeGraphEvent.PreviousEventData.Alias) {
			message = message + "Alias changed.\n"
		}
		// TODO FIXME fix sorting of the data in Features and Addresses before comparing
		//if nodeGraphEvent.Features != nodeGraphEvent.PreviousEventData.Features {
		//	message = message + "Features changed.\n"
		//}
		//if nodeGraphEvent.Addresses != nodeGraphEvent.PreviousEventData.Addresses {
		//	message = message + "Addresses changed.\n"
		//}
		if message != "" {
			sendBotMessages(message, communications)
		}
	}
}

func compareGraphSyncTime(previousInformation commons.InformationResponse,
	newInformation commons.InformationResponse,
	graphInSyncTime map[int]time.Time,
	torqNodeSettings commons.ManagedNodeSettings,
	message string) string {

	if newInformation.GraphSynced {
		graphInSyncTime[torqNodeSettings.NodeId] = time.Now()
		if !previousInformation.GraphSynced {
			message = message + "Graph is synced\n"
		}
	}
	lastGraphInSyncTime, lastGraphInSyncTimeExists := graphInSyncTime[torqNodeSettings.NodeId]
	if lastGraphInSyncTimeExists && int(time.Since(lastGraphInSyncTime).Seconds()) > 60 {
		message = message + "Graph is out of sync\n"
	}
	return message
}

func compareChainSyncTime(previousInformation commons.InformationResponse,
	newInformation commons.InformationResponse,
	chainInSyncTime map[int]time.Time,
	torqNodeSettings commons.ManagedNodeSettings,
	message string) string {

	if newInformation.ChainSynced {
		chainInSyncTime[torqNodeSettings.NodeId] = time.Now()
		if !previousInformation.ChainSynced {
			message = message + "Chain is synced\n"
		}
	}
	lastChainInSyncTime, lastChainInSyncTimeExists := chainInSyncTime[torqNodeSettings.NodeId]
	if lastChainInSyncTimeExists && int(time.Since(lastChainInSyncTime).Seconds()) > 60 {
		message = message + "Chain is out of sync\n"
	}
	return message
}

func comparePendingChannelCount(previousInformation commons.InformationResponse,
	newInformation commons.InformationResponse,
	message string) string {

	if previousInformation.PendingChannelCount != newInformation.PendingChannelCount {
		if newInformation.PendingChannelCount == 0 {
			message = message + "No pending channels anymore\n"
		} else {
			message = message + fmt.Sprintf("Pending channels: %v -> %v\n",
				previousInformation.PendingChannelCount, newInformation.PendingChannelCount)
		}
	}
	return message
}

func compareInactiveChannelCount(previousInformation commons.InformationResponse,
	newInformation commons.InformationResponse,
	message string) string {

	if previousInformation.InactiveChannelCount != newInformation.InactiveChannelCount {
		if newInformation.InactiveChannelCount == 0 {
			message = message + "No inactive channels anymore\n"
		} else {
			message = message + fmt.Sprintf("Inactive channels: %v -> %v\n",
				previousInformation.InactiveChannelCount, newInformation.InactiveChannelCount)
		}
	}
	return message
}

func compareVersion(previousInformation commons.InformationResponse,
	newInformation commons.InformationResponse,
	message string) string {

	if previousInformation.Version != newInformation.Version {
		message = message + fmt.Sprintf("Version changed from %v to %v\n",
			previousInformation.Version, newInformation.Version)
	}
	return message
}

func compareActiveChannelCount(previousInformation commons.InformationResponse,
	newInformation commons.InformationResponse,
	message string) string {

	if previousInformation.ActiveChannelCount != newInformation.ActiveChannelCount {
		if newInformation.ActiveChannelCount == 0 {
			message = message + "No active channels anymore\n"
		} else {
			message = message + fmt.Sprintf("Active channels: %v -> %v\n",
				previousInformation.ActiveChannelCount, newInformation.ActiveChannelCount)
		}
	}
	return message
}

func sendBotMessages(communicationMessage string, communicationDestinations []Communication) {
	for _, communication := range communicationDestinations {
		log.Info().Msgf("Notifier sending telegram communication: %v", communicationMessage)
		switch communication.TargetType {
		case CommunicationTelegramHighPriority:
			log.Info().Msgf("Notifier sending HIGH priority telegram communication (%v): %v", communication.TargetName, communicationMessage)
			bot := MessageForBot{
				Message: communicationMessage,
				Telegram: MessageForTelegram{
					Id: communication.TargetNumber,
				},
			}
			SendTelegramBotMessages(bot, CommunicationTelegramHighPriority)
		case CommunicationTelegramLowPriority:
			log.Info().Msgf("Notifier sending LOW priority telegram communication (%v): %v", communication.TargetName, communicationMessage)
			bot := MessageForBot{
				Message: communicationMessage,
				Telegram: MessageForTelegram{
					Id: communication.TargetNumber,
				},
			}
			SendTelegramBotMessages(bot, CommunicationTelegramLowPriority)
		case CommunicationSlack:
			log.Info().Msgf("Notifier sending Slack communication (%v): %v", communication.TargetName, communicationMessage)
			SendSlackBotMessages(MessageForBot{
				Message: communicationMessage,
				Slack: MessageForSlack{
					Channel: communication.TargetText,
					Color:   "#283B4C",
				},
			})
		}
	}
}

func HandleMessage(db *sqlx.DB,
	messageForBot MessageForBot,
	publicKeyFromChannel string,
	commandFromChannel string,
	communicationTargetType CommunicationTargetType) {

	switch commandFromChannel {
	case VectorButton:
		fallthrough
	case StatusButton:
		messageForBot = processStatusRequest(db, communicationTargetType, publicKeyFromChannel, messageForBot)
	case SettingsButton:
		if messageForBot.IsTelegram() {
			SendNodeSettingsMenu(db, communicationTargetType, messageForBot)
			//telegram.SendChannelSettingsMenu(db, userId, communicationTargetType, messageForBot)
		}
	//case PingButton:
	//	messageForBot = processPingRequest(db, communicationTargetType, messageFromChannel, messageForBot)
	case RegisterButton:
		messageForBot = processRegisterRequest(db, communicationTargetType, commons.GetActiveTorqNodeSettings(),
			messageForBot)
	case UnregisterButton:
		messageForBot = processUnregisterRequest(db, communicationTargetType, messageForBot)
	case PublicKeyButton:
		PublicKeys[communicationTargetType][messageForBot.GetChannelIdentifier()] = publicKeyFromChannel
		//messageForBot.Menus = []Menu{MenuNodeSettings}
		//if messageForBot.IsTelegram() {
		//telegram.SendChannelSettingsMenu(db, userId, communicationTargetType, messageForBot)
		//}
	//case ChannelIdButton:
	//	ChannelIds[communicationTargetType][messageForBot.GetChannelIdentifier()] = messageFromChannel
	//	sendNodeSettingsMenu(ctx, messageForBot.Chat.ID, db, communicationTargetType, highPriority, bot, messageForBot)
	//	sendChannelSettingsMenu(ctx, messageForBot.Chat.ID, db, communicationTargetType, highPriority, bot, messageForBot)
	case MenuButton:
		fallthrough
	default:
		messageForBot.Menus = []Menu{MenuMain}
	}
	if messageForBot.HasMessage() || messageForBot.HasMenu() {
		switch communicationTargetType {
		case CommunicationSlack:
			SendSlackBotMessages(messageForBot)
		case CommunicationTelegramHighPriority, CommunicationTelegramLowPriority:
			SendTelegramBotMessages(messageForBot, communicationTargetType)
		}
	}
}

func HandleButton(db *sqlx.DB,
	messageForBot MessageForBot,
	commandFromChannel string,
	messageFromChannel string,
	communicationTargetType CommunicationTargetType) {

	communicationIds := getCommunicationIds(db, messageForBot)
	switch commandFromChannel {
	case VectorButton:
		fallthrough
	case StatusButton:
		messageForBot = processStatusRequest(db, communicationTargetType, messageFromChannel, messageForBot)
	case SettingsButton:
		messageForBot = processSettingsRequest(db, communicationIds, communicationTargetType, messageFromChannel, messageForBot)
	//case PingButton:
	//	messageForBot = processPingRequest(db, communicationTargetType, messageFromChannel, messageForBot)
	case RegisterButton:
		messageForBot = processRegisterRequest(db, communicationTargetType, commons.GetActiveTorqNodeSettings(), messageForBot)
	case UnregisterButton:
		messageForBot = processUnregisterRequest(db, communicationTargetType, messageForBot)
	default:
		messageForBot = processSettingsRequest(db, communicationIds, communicationTargetType, messageFromChannel, messageForBot)
	}
	if messageForBot.HasMessage() || messageForBot.HasMenu() {
		switch communicationTargetType {
		case CommunicationSlack:
			SendSlackBotMessages(messageForBot)
		case CommunicationTelegramHighPriority, CommunicationTelegramLowPriority:
			SendTelegramBotMessages(messageForBot, communicationTargetType)
		}
	}
}

func processSettingsRequest(db *sqlx.DB,
	communicationIds []int,
	communicationTargetType CommunicationTargetType,
	settings string,
	messageForBot MessageForBot) MessageForBot {

	nodeSettings := map[string]bool{
		"nodeDetailsDeactivate": true,
		"nodeDetailsActivate":   true,
		//"nodeStatusActivate":    true,
		//"nodeHtlcDeactivate":        true,
		//"nodeHtlcActivate":          true,
		//"nodeTimeLockDeltactivate":  true,
		//"nodeTimeLockDeltaActivate": true,
		//"nodeFeeDeactivate":         true,
		//"nodeFeeActivate":           true,
		//"nodeFeeMisMatchDeactivate": true,
		//"nodeFeeMisMatchActivate":   true,
	}

	if settings == "" {
		SendNodeSettingsMenu(db, communicationTargetType, messageForBot)
		return messageForBot
	}
	var channelIds []int
	var nodeIds []int
	var err error
	if _, exists := nodeSettings[settings]; exists {
		nodeIds, err = GetNodeIdsByCommunication(db, communicationTargetType)
		cachedPublicKey := PublicKeys[communicationTargetType][messageForBot.GetChannelIdentifier()]
		if PublicKeys[communicationTargetType][messageForBot.GetChannelIdentifier()] != "" {
			nodeId := commons.GetNodeIdByPublicKey(cachedPublicKey, commons.Bitcoin, commons.MainNet)
			nodeIds = []int{nodeId}
			if nodeId == 0 {
				nodeIds = []int{}
			}
		}
		if err != nil {
			log.Error().Err(err).Msgf(
				"Failed to obtain nodeIds with communicationTargetType: %v", communicationTargetType)
			messageForBot.Message = "We could not find existing node."
			messageForBot.Error = err.Error()
			return messageForBot
		}
		//	channelIds, err = channels.GetChannelIdsByChatId(db, botMessage.Chat.ID, communicationTargetType, channelId)
	}

	if len(channelIds) != 1 && len(nodeIds) != 1 {
		messageForBot.Message = "We could not find your node/channel.\nMake sure it is correctly registered and referenced."
		messageForBot.Error = err.Error()
		return messageForBot
	}
	channelId := 0
	if len(channelIds) == 1 {
		channelId = channelIds[0]
	}
	nodeId := 0
	if len(nodeIds) == 1 {
		nodeId = nodeIds[0]
	}
	activate := strings.Contains(settings, "Activate")
	communicationType := GetCommunicationType(settings)
	var communications []Communication
	if channelId != 0 {
		communications, err = GetCommunicationsByChannelIdAndTargetTypes(db, nodeId, channelId, communicationTargetType)
	} else if nodeId != 0 {
		communications, err = GetCommunicationsByNodeIdAndTargetTypes(db, nodeId, communicationTargetType)
	}
	if err != nil {
		log.Error().Err(err).Msgf(
			"Failed to obtain communication from: %v parameter: %v",
			messageForBot.GetChannelIdentifier(), settings)
		messageForBot.Message = "We could not find existing settings."
		messageForBot.Error = err.Error()
		return messageForBot
	}
	if communicationType == nil {
		messageForBot.Message = "We could not parse the option value.\nMake sure you choose an existing option.\n"
		messageForBot.Error = err.Error()
		return messageForBot
	}
	for _, communication := range communications {
		// User requesting to activate but it's already active
		if activate && communication.HasCommunicationType(*communicationType) {
			// So nothing to do here
			continue
		}
		// User requesting to inactivate but it's already inactive
		if !activate && !communication.HasCommunicationType(*communicationType) {
			// So nothing to do here
			continue
		}
		if activate {
			communication.AddCommunicationType(*communicationType)
		} else {
			communication.RemoveCommunicationType(*communicationType)
		}
		_, err = SetCommunication(db, communication)
		if err != nil {
			log.Error().Err(err).Msgf(
				"Failed to persist communication fro: %v parameter: %v",
				messageForBot.GetChannelIdentifier(), settings)
			messageForBot.Message = "We could store the settings."
			messageForBot.Error = err.Error()
			return messageForBot
		}
	}
	messageForBot.Message = "Setting registered."
	return messageForBot
}

func processRegisterRequest(db *sqlx.DB,
	communicationTargetType CommunicationTargetType,
	activeTorqNodeSettings []commons.ManagedNodeSettings,
	messageForBot MessageForBot) MessageForBot {

	existingNodeIds, err := GetNodeIdsByCommunication(db, communicationTargetType)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to check for existing nodes")
		messageForBot.Message = "Something went wrong verifying existing configurations."
		messageForBot.Error = err.Error()
		return messageForBot
	}
	var unregisteredNodeIds []int
	for _, activeTorqNodeSetting := range activeTorqNodeSettings {
		if slices.Contains(existingNodeIds, activeTorqNodeSetting.NodeId) {
			continue
		}
		unregisteredNodeIds = append(unregisteredNodeIds, activeTorqNodeSetting.NodeId)
	}
	if len(unregisteredNodeIds) == 0 {
		messageForBot.Message = "All nodes are already registered."
		return messageForBot
	}
	for _, unregisteredNodeId := range unregisteredNodeIds {
		communication := Communication{
			TargetType: communicationTargetType,
			NodeId:     unregisteredNodeId,
		}
		if messageForBot.IsSlack() {
			communication.TargetName = messageForBot.Slack.ReplyTo
			communication.TargetText = messageForBot.Slack.Channel
		}
		if messageForBot.IsTelegram() {
			communication.TargetName = messageForBot.Telegram.UserName
			communication.TargetNumber = messageForBot.Telegram.Id
		}
		communication.AddCommunicationType(NodeDetailsChanged)
		_, err = AddCommunication(db, communication)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to add communication: %v", communication)
			messageForBot.Message = "Something went wrong (code: ac)."
			messageForBot.Error = err.Error()
			return messageForBot
		}
	}
	messageForBot.Message = fmt.Sprintf("%v node(s) registered.", len(unregisteredNodeIds))
	return messageForBot
}

func processUnregisterRequest(db *sqlx.DB,
	communicationTargetType CommunicationTargetType,
	messageForBot MessageForBot) MessageForBot {

	existingNodeIds, err := GetNodeIdsByCommunication(db, communicationTargetType)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to check for existing nodes")
		messageForBot.Message = "Something went wrong verifying existing configurations."
		messageForBot.Error = err.Error()
		return messageForBot
	}
	if len(existingNodeIds) == 0 {
		messageForBot.Message = "No registration found."
		return messageForBot
	}
	for _, existingNodeId := range existingNodeIds {
		if messageForBot.IsSlack() {
			_, err = RemoveCommunicationByTargetText(db, existingNodeId, communicationTargetType, messageForBot.Slack.Channel)
		}
		if messageForBot.IsTelegram() {
			_, err = RemoveCommunicationByTargetNumber(db, existingNodeId, communicationTargetType, messageForBot.Telegram.Id)
		}
		if err != nil {
			log.Error().Err(err).Msgf("Failed to remove communication for nodeId: %v", existingNodeId)
			messageForBot.Message = "Something went wrong (code: rc)."
			messageForBot.Error = err.Error()
			return messageForBot
		}
	}
	messageForBot.Message = fmt.Sprintf("%v node(s) unregistered.", len(existingNodeIds))
	return messageForBot
}

func processStatusRequest(db *sqlx.DB,
	communicationTargetType CommunicationTargetType,
	publicKey string,
	messageForBot MessageForBot) MessageForBot {

	var inputText string
	var nodeIds []int
	nodeIds, messageForBot = getNodeIds(db, communicationTargetType, publicKey, messageForBot)
	if messageForBot.HasMessage() {
		return messageForBot
	}
	for _, nodeId := range nodeIds {
		information, err := lightning.GetInformationRequest(nodeId)
		if err != nil {
			log.Error().Err(err).Msgf(
				"Failed to obtain node information from: %v or publicKey: %v", messageForBot.GetChannelIdentifier(), publicKey)
			messageForBot.Message = "Something went wrong (gnvpe)."
			messageForBot.Error = err.Error()
			return messageForBot
		}
		inputText = inputText + "Block Height: " + strconv.Itoa(int(information.BlockHeight)) + "\n" +
			"Channels active: " + strconv.Itoa(information.ActiveChannelCount) +
			", inactive: " + strconv.Itoa(information.InactiveChannelCount) +
			", pending: " + strconv.Itoa(information.PendingChannelCount) + "\n"
		if information.GraphSynced {
			inputText = inputText + "Graph is synced\n"
		} else {
			inputText = inputText + "Graph is out of sync\n"
		}
		if information.ChainSynced {
			inputText = inputText + "Chain is synced\n"
		} else {
			inputText = inputText + "Chain is out of sync\n"
		}
		inputText = inputText + "LND Version: " + information.Version + "\n" +
			"Torq Version: " + build.ExtendedVersion() + "\n"
		inputText = inputText + "\n\n"
	}
	messageForBot.Message = inputText
	return messageForBot
}

func getNodeIds(db *sqlx.DB,
	communicationTargetType CommunicationTargetType,
	publicKey string,
	messageForBot MessageForBot) ([]int, MessageForBot) {

	nodeIds, err := GetNodeIdsByCommunication(db, communicationTargetType)
	if publicKey != "" {
		nodeId := commons.GetNodeIdByPublicKey(publicKey, commons.Bitcoin, commons.MainNet)
		if slices.Contains(nodeIds, nodeId) {
			nodeIds = []int{nodeId}
		} else {
			nodeIds = []int{}
		}
	}
	if err != nil {
		log.Error().Err(err).Msgf(
			"Failed to obtain nodeId from: %v or publicKey: %v", messageForBot.GetChannelIdentifier(), publicKey)
		messageForBot.Message = "Something went wrong (nibcd)."
		messageForBot.Error = err.Error()
		return nil, messageForBot
	}
	if len(nodeIds) == 0 {
		messageForBot.Message = "/register > Node Registration"
		return nodeIds, messageForBot
	}
	return nodeIds, messageForBot
}

//func processPingRequest(db *sqlx.DB,
//	communicationTargetType CommunicationTargetType,
//	publicKey string,
//	messageForBot MessageForBot) MessageForBot {
//
//	var nodeIds []int
//	nodeIds, messageForBot = getNodeIds(db, communicationTargetType, publicKey, messageForBot)
//	if len(nodeIds) > 1 {
//		messageForBot.Message = "Please specify node via $PUBLICKEY parameter"
//		return messageForBot
//	}
//	nodeInformation, connectionDuration, err := lnd.GetNodeInfoAndConnect(nodeIds[0], true)
//	if err != nil {
//		log.Error().Err(err).Msgf("Failed to get node information for nodeId: %v", nodeIds[0])
//		messageForBot.Message = "We could not connect to your node.\nMake sure it is correctly registered and referenced."
//		messageForBot.Error = err.Error()
//		return messageForBot
//	}
//	p := message.NewPrinter(language.English)
//	inputText := fmt.Sprintf("%v (%v)\nChannels: %v\nCapacity: %v sats (%v btc)\n",
//		nodeInformation.Node.Alias, nodeInformation.Node.Color,
//		p.Sprintf("%d", nodeInformation.NumChannels),
//		p.Sprintf("%d", nodeInformation.TotalCapacity),
//		p.Sprintf("%d", nodeInformation.TotalCapacity/100_000_000))
//	if nodeInformation.Node.Addresses != nil && len(nodeInformation.Node.Addresses) > 0 {
//		for _, address := range nodeInformation.Node.Addresses {
//			inputText = inputText + fmt.Sprintf("Address (%v): %v\n", address.Network, address.Addr)
//		}
//	}
//	messageForBot.Message = inputText + fmt.Sprintf("Time to connect: %s", connectionDuration)
//	return messageForBot
//}

func getCommunicationIds(db *sqlx.DB, messageForBot MessageForBot) []int {
	if messageForBot.IsSlack() {
		communicationIds, err := GetCommunicationIdsByTargetText(db, messageForBot.Slack.Channel)
		if err != nil {
			log.Error().Err(err).Msgf("Slack bot Get communicationIds By Channel failed channel: %v (%v)", messageForBot.Slack.Channel, messageForBot.Slack.ReplyTo)
		}
		return communicationIds
	}
	if messageForBot.IsTelegram() {
		communicationIds, err := GetCommunicationIdsByTargetNumber(db, messageForBot.Telegram.Id)
		if err != nil {
			log.Error().Err(err).Msgf("Telegram bot Get communicationIds By ChatId failed chatId: %v (%v)", messageForBot.Telegram.Id, messageForBot.Telegram.ReplyToMessageId)
		}
		return communicationIds
	}
	return nil
}
