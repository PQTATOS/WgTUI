package tgvpnbot

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/PQTATOS/WireTgBot/pkg/telegram"
	"github.com/PQTATOS/WireTgBot/pkg/vpnbot"
	"github.com/PQTATOS/WireTgBot/pkg/vpn"
)

type Processor struct {
	TgClient clients.BotClient
	WireguardClient vpn.VPNServer
	updateId int
	handlers map[string]func(vpnbot.Event, vpnbot.Processor)
}

type Meta struct {
	Username string
	ChatId int
	UpdateID int
}

func New(bot clients.BotClient, vpn vpn.VPNServer) *Processor {
	return &Processor{
		TgClient: bot,
		WireguardClient: vpn,
		updateId: -1,
		handlers: make(map[string]func(vpnbot.Event, vpnbot.Processor)),
	}
}

func (proc *Processor) Process(event vpnbot.Event) error {
	proc.updateId = event.Meta.(Meta).UpdateID + 1
	msg := event.Text
	handler, ok := proc.handlers[msg]
	if !ok {
		return fmt.Errorf("Process: handler doesnt exist")
	}
	go handler(event, proc)
	return nil
}

func (proc *Processor) AddHandler(template string, handle func(vpnbot.Event, vpnbot.Processor)) error {
	proc.handlers[template] = handle
	return nil
}

func (proc *Processor) Start() {
	pollsWithoutUpdates := 0
	for {
		events := proc.fetch()
		slog.Info(fmt.Sprintf("Processor recive %d updates", len(events)))

		for i := range events {
			proc.Process(events[i])
		}

		if len(events) == 0 {
			pollsWithoutUpdates++
			if pollsWithoutUpdates > 4 {
				slog.Info("Processor is sleeping")
				time.Sleep(1 * time.Minute)
			}
		} else {
			pollsWithoutUpdates = 0
		}
	}
}

func (proc *Processor) fetch() []vpnbot.Event {
	updates, _ := proc.TgClient.GetUpdates(60, proc.updateId)
	events := make([]vpnbot.Event, len(updates))
	for i := range updates {
		events[i] = updToEvent(updates[i])
	}
	return events

}

func updToEvent(upd clients.Update) vpnbot.Event {
	return vpnbot.Event{
		Text: upd.Message.Text,
		Meta: Meta{
			Username: *upd.Message.Chat.Username,
			ChatId: upd.Message.Chat.Id,
			UpdateID: upd.UpdateId,
		},
	}
}