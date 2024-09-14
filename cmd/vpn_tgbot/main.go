package main

import (
	"log/slog"
	"github.com/PQTATOS/WireTgBot/pkg/telegram/tgclient"
	"github.com/PQTATOS/WireTgBot/pkg/vpnbot"
	"github.com/PQTATOS/WireTgBot/pkg/vpnbot/tgvpnbot"
	"github.com/PQTATOS/WireTgBot/pkg/vpn/testvpn"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Info("Creating bot")

	vpnTgbot := tgvpnbot.New(
		tgclient.New(""),
		testvpn.New("tmpTest"),
	)
	vpnTgbot.AddHandler("/Text", textTest)
	vpnTgbot.AddHandler("/Photo", textPhoto)
	vpnTgbot.AddHandler("/Document", textDocs)
	
	slog.Info("Starting bot")
	vpnTgbot.Start()
}

func textTest(event vpnbot.Event, bot vpnbot.Processor) {
	bot.(*tgvpnbot.Processor).TgClient.SendMessage(event.Meta.(tgvpnbot.Meta).ChatId, "TextTest")
}

func textDocs(event vpnbot.Event, bot vpnbot.Processor) {
	cfg, _ := bot.(*tgvpnbot.Processor).WireguardClient.NewConfig(event.Meta.(tgvpnbot.Meta).Username)
	bot.(*tgvpnbot.Processor).TgClient.SendDocument(event.Meta.(tgvpnbot.Meta).ChatId, cfg.Meta.(string))
}

func textPhoto(event vpnbot.Event, bot vpnbot.Processor) {
	bot.(*tgvpnbot.Processor).TgClient.SendPhoto(event.Meta.(tgvpnbot.Meta).ChatId, "test.jpg")
}