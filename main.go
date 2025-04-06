package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jamesmcdonald/plexer/internal/plex"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Ping the bot, and get a pong back",
	},
	{
		Name:        "plexlibraries",
		Description: "List the Plex libraries",
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		user := i.User
		if user == nil {
			user = i.Member.User
		}
		slog.Info("Ponging back", "user", user.Username)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "pong",
			},
		})
	},
}

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
	plexToken := os.Getenv("PLEX_TOKEN")
	guildID := flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	unregister := flag.Bool("unregister", false, "Unregister commands")
	plexEndpoint := flag.String("plex", "", "Plex server endpoint")
	flag.Parse()

	if *plexEndpoint == "" {
		slog.Error("Plex server endpoint is not set")
		os.Exit(1)
	}

	plex := plex.New(*plexEndpoint, plexToken)
	commandHandlers["plexlibraries"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		slog.Info("Getting Plex libraries")
		libraries, err := plex.GetLibraries()
		if err != nil {
			slog.Error("Cannot get libraries", "error", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Cannot get libraries",
				},
			})
			return
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Plex libraries: " + strings.Join(libraries, ", "),
			},
		})
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	dg.AddHandler(pong)
	dg.AddHandler(func(dg *discordgo.Session, r *discordgo.Ready) {
		slog.Info("Logged in", "user", dg.State.User.Username+"#"+dg.State.User.Discriminator)
	})

	dg.AddHandler(func(dg *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(dg, i)
		}
	})

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	err = dg.Open()
	if err != nil {
		panic(err)
	}
	defer dg.Close()

	slog.Info("Registering commands")
	appCommands, err := dg.ApplicationCommands(dg.State.User.ID, *guildID)
	appCommandNames := make(map[string]int, len(appCommands))
	if err != nil {
		slog.Error("Cannot get commands", "error", err)
		os.Exit(1)
	}
	for i, acmd := range appCommands {
		appCommandNames[acmd.Name] = i
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		if _, ok := appCommandNames[v.Name]; ok {
			slog.Info("Command already registered", "name", v.Name)
			registeredCommands[i] = appCommands[appCommandNames[v.Name]]
			continue
		}
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, *guildID, v)
		if err != nil {
			slog.Error("Cannot create command", "command", v.Name, "error", err)
		}
		registeredCommands[i] = cmd
		slog.Info("Command registered", "name", cmd.Name, "id", cmd.ID)
	}

	fmt.Println("Plexer is running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if !*unregister {
		return
	}

	slog.Info("Unregistering commands")
	for _, v := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, *guildID, v.ID)
		if err != nil {
			slog.Error("Cannot delete command", "name", v.Name, "error", err)
		}
	}
}
