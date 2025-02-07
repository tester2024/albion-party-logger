package main

import (
	"context"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

type BotContact struct {
	m           *GameDataManager
	partyLogger webhook.Client
	lootLogger  webhook.Client
}

func (l *BotContact) Close(ctx context.Context) {
	if l.lootLogger != nil {
		l.lootLogger.Close(ctx)
	}

	if l.partyLogger != nil {
		l.partyLogger.Close(ctx)
	}
}

func (l *BotContact) LogParty(message discord.WebhookMessageCreate) {
	if l.partyLogger == nil {
		return
	}

	_, err := l.partyLogger.CreateMessage(message)
	if err != nil {
		panic(err)
	}
}

func (l *BotContact) LogPartyAdd(uid uuid.UUID) {
	username := l.m.GetUsername(uid)
	partyOwnerName := l.m.GetUsername(l.m.CurrentParty.PartyOwner)
	selfUsername := l.m.GetSelfUsername()

	l.LogParty(discord.NewWebhookMessageCreateBuilder().
		AddEmbeds(
			discord.NewEmbedBuilder().
				SetTitlef("✅ **Member Joined** - %s", partyOwnerName).
				SetFooter(selfUsername, "").
				SetColor(0x00ff00).
				SetDescriptionf("👤 **Username:** %s", username).
				SetTimestamp(time.Now()).
				Build(),
		).
		Build(),
	)
}

func (l *BotContact) LogPartyLeft(uid uuid.UUID) {
	username := l.m.GetUsername(uid)
	partyOwnerName := l.m.GetUsername(l.m.CurrentParty.PartyOwner)
	selfUsername := l.m.GetSelfUsername()

	l.LogParty(discord.NewWebhookMessageCreateBuilder().
		AddEmbeds(
			discord.NewEmbedBuilder().
				SetTitlef("❌ **Member Left** - %s", partyOwnerName).
				SetFooter(selfUsername, "").
				SetColor(0xff0000).
				SetDescriptionf("👤 **Username:** %s", username).
				SetTimestamp(time.Now()).
				Build(),
		).
		Build(),
	)
}

func (l *BotContact) LogSelfLeave(uid uuid.UUID) {
	username := l.m.GetUsername(uid)
	partyOwnerName := l.m.GetUsername(l.m.CurrentParty.PartyOwner)
	selfUsername := l.m.GetSelfUsername()

	l.LogParty(discord.NewWebhookMessageCreateBuilder().
		AddEmbeds(
			discord.NewEmbedBuilder().
				SetTitlef("🚪 **I Leave Party** - %s", partyOwnerName).
				SetFooter(selfUsername, "").
				SetColor(0xffa500).
				SetDescriptionf("👤 **Username:** %s", username).
				SetTimestamp(time.Now()).
				Build(),
			l.generateHistoryEmbed(),
		).
		Build(),
	)
}

func (l *BotContact) LogPartyUpdate(removedPlayers []uuid.UUID, addedPlayers []uuid.UUID) {
	if len(removedPlayers) == 0 && len(addedPlayers) == 0 {
		return
	}

	partyOwnerName := l.m.GetUsername(l.m.CurrentParty.PartyOwner)
	selfUsername := l.m.GetSelfUsername()

	embedBuilder := discord.NewEmbedBuilder().
		SetTitlef("🔄 **Party Update** - %s", partyOwnerName).
		SetFooter(selfUsername, "").
		SetColor(0x3498db).
		SetTimestamp(time.Now())

	if len(addedPlayers) > 0 {
		var addedPlayersDesc string
		for _, player := range addedPlayers {
			username := l.m.GetUsername(player)
			if username == "" {
				username = player.String()
			}
			addedPlayersDesc += "✅ " + username + "\n"
		}
		embedBuilder.AddField("**Added Players:**", addedPlayersDesc, false)
	}

	if len(removedPlayers) > 0 {
		var removedPlayersDesc string
		for _, player := range removedPlayers {
			username := l.m.GetUsername(player)
			if username == "" {
				username = player.String()
			}
			removedPlayersDesc += "❌ " + username + "\n"
		}
		embedBuilder.AddField("**Removed Players:**", removedPlayersDesc, false)
	}

	l.LogParty(discord.NewWebhookMessageCreateBuilder().
		AddEmbeds(
			embedBuilder.Build(),
			l.generateHistoryEmbed(),
		).
		Build(),
	)
}

func (l *BotContact) LogPartyCreate(party *Party) {
	partyOwnerName := l.m.GetUsername(l.m.CurrentParty.PartyOwner)
	selfUsername := l.m.GetSelfUsername()

	members := new(strings.Builder)
	for _, u := range party.Members.Values() {
		username := l.m.GetUsername(u)
		members.WriteString("👤 ")
		members.WriteString(username)
		members.WriteRune('\n')
	}

	l.LogParty(discord.NewWebhookMessageCreateBuilder().
		AddEmbeds(
			discord.NewEmbedBuilder().
				SetTitlef("🎉 **New Party Created** - %s", partyOwnerName).
				AddField("**Members:**", members.String(), false).
				SetFooter(selfUsername, "").
				SetColor(0x00ff00).
				SetTimestamp(time.Now()).
				Build(),
		).
		Build(),
	)
}

func (l *BotContact) LogPartyDisband() {
	partyOwnerName := l.m.GetUsername(l.m.CurrentParty.PartyOwner)
	selfUsername := l.m.GetSelfUsername()

	l.LogParty(discord.NewWebhookMessageCreateBuilder().
		AddEmbeds(
			discord.NewEmbedBuilder().
				SetTitlef("💔 **Party Disbanded** - %s", partyOwnerName).
				SetFooter(selfUsername, "").
				SetColor(0xff0000).
				SetTimestamp(time.Now()).
				Build(),
			l.generateHistoryEmbed(),
		).
		Build(),
	)
}

func (l *BotContact) generateHistoryEmbed() discord.Embed {
	builder := new(strings.Builder)
	for _, entry := range l.m.CurrentParty.History {
		builder.WriteString("🕒 ")
		builder.WriteString(discord.FormattedTimestampMention(entry.Timestamp.Unix(), discord.TimestampStyleRelative))
		builder.WriteString(" - ")
		builder.WriteString(l.m.GetUsername(entry.User))
		builder.WriteString(" ➡ ")
		builder.WriteString(entry.Action.String())
		builder.WriteString("\n")
	}

	return discord.NewEmbedBuilder().
		SetTitlef("📜 **Party History** - %s", l.m.GetUsername(l.m.CurrentParty.PartyOwner)).
		SetFooter(l.m.GetSelfUsername(), "").
		SetColor(0x7289da).
		SetDescription(builder.String()).
		SetTimestamp(time.Now()).
		Build()
}

func NewDiscordLogger(m *GameDataManager, partyLoggerUrl string, lootLoggerUrl string) *BotContact {
	var partyLogger, lootLogger webhook.Client
	var err error

	partyLogger, err = webhook.NewWithURL(partyLoggerUrl)
	if err != nil {
		log.Println(err)
	}

	lootLogger, err = webhook.NewWithURL(lootLoggerUrl)
	if err != nil {
		log.Println(err)
	}

	return &BotContact{
		m:           m,
		partyLogger: partyLogger,
		lootLogger:  lootLogger,
	}
}
