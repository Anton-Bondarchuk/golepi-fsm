package models

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Telegram      *tgbotapi.BotAPI
	UpdateChannel tgbotapi.UpdatesChannel
}
