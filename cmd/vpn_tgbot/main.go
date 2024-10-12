package main

import (
	"log/slog"
	"os"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/PQTATOS/WireTgBot/pkg/config"
	"github.com/PQTATOS/WireTgBot/pkg/telegram/tgclient"
	"github.com/PQTATOS/WireTgBot/pkg/vpn/wgvpn"
	"github.com/PQTATOS/WireTgBot/pkg/vpnbot"
	"github.com/PQTATOS/WireTgBot/pkg/vpnbot/tgvpnbot"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	slog.Info("Cheling Packages")
	wgVpn, err := wgvpn.New(wgvpn.CreateConfigLocalRepo("/etc/wireguard"))
	if err != nil {
		slog.Error("Packeges not installed: %s", slog.Any("%s", err))
		os.Exit(1)
	}
	slog.Info("Init wireguard")
	if err := wgVpn.Init(); err != nil {
		slog.Error("Cant init wireguard: %s", slog.Any("%s", err))
		os.Exit(1)
	}
	slog.Info("Reading config file")
	if err := config.ReadConfig(); err != nil {
		slog.Error("Config error: %s", slog.Any("%s", err))
		os.Exit(1)
	}
	slog.Info("Creating bot")
	vpnTgbot := tgvpnbot.New(
		tgclient.New(config.Cfg.Bot.Token),
		wgVpn,
	)
	vpnTgbot.AddHandler("/Text", textTest)
	vpnTgbot.AddHandler("/Config", testConfig)
	
	slog.Info("Starting bot")
	vpnTgbot.Start()
}

func textTest(event vpnbot.Event, bot vpnbot.Processor) {
	bot.(*tgvpnbot.Processor).TgClient.SendMessage(event.Meta.(tgvpnbot.Meta).ChatId, "TextTest")
}

func testConfig(event vpnbot.Event, bot vpnbot.Processor) {
	vpnbot := bot.(*tgvpnbot.Processor)
	meta := event.Meta.(tgvpnbot.Meta)
	cfg, err := vpnbot.WireguardClient.NewConfig(meta.Username)
	if err != nil {
		slog.Error("Cant create client config file:", err)
		return
	}
	slog.Debug(string(cfg.Data))
	vpnbot.TgClient.SendDocument(meta.ChatId, cfg.FilePath)

	qr, _  := qrcode.Encode(string(cfg.Data), qrcode.Medium, 256)
	vpnbot.TgClient.SendPhotoRaw(meta.ChatId, qr)
}