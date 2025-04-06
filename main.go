package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func pong(s *discordgo.Session, m *discordgo.MessageCreate) {
	slog.Info("Message received", "user", m.Author.Username, "content", m.Content)
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.Contains(m.Content, "ping") {
		slog.Info("Ponging back", "user", m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, "pong")
	}
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	dg.AddHandler(pong)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	err = dg.Open()
	if err != nil {
		panic(err)
	}
	defer dg.Close()

	fmt.Println("Plexer is running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
