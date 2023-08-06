package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v53/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	Token       string
	ClientID    string
	ChannelID   string
	GithubToken string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Token = os.Getenv("DISCORD_TOKEN")
	ClientID = os.Getenv("DISCORD_CLIENT_ID")
	ChannelID = os.Getenv("DISCORD_CHANNEL_ID")
	GithubToken = os.Getenv("GITHUB_ACCESS_TOKEN")
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		fmt.Println(err)
	}

	go runEvery(time.Second*30, periodicTask, dg)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func runEvery(d time.Duration, f func(*discordgo.Session), s *discordgo.Session) {
	for range time.Tick(d) {
		f(s)
	}
}

func periodicTask(s *discordgo.Session) {
	fmt.Println("Task is running...")
	s.ChannelMessageSend(ChannelID, getMostRecentCommits())
}

func getMostRecentCommits() string {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		log.Fatalf("Error fetching user: %v", err)
	}

	return fmt.Sprintf("Hello, %s\n", *user.Name)
}
