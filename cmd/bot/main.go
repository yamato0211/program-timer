package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
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

	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		fmt.Println(err)
	}

	go runEvery(time.Second*15, periodicTask, dg)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func runEvery(d time.Duration, f func(*discordgo.Session), s *discordgo.Session) {
	for range time.Tick(d) {
		f(s)
	}
}

func periodicTask(s *discordgo.Session) {
	fmt.Println("Task is running...")
	latestCommit := getMostRecentCommits()
	if isMoreThan24HoursAgo(latestCommit.Commit.Author.Date.Time) {
		s.ChannelMessageSend(ChannelID, "@everyone コード書けよ!!")
	}
}

func getMostRecentCommits() *github.RepositoryCommit {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Fatalf("Error fetching user: %v", err)
	}
	username := *user.Login

	opt := &github.RepositoryListOptions{
		Affiliation: "owner",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	repos, _, err := client.Repositories.List(ctx, username, opt)
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].GetPushedAt().Time.After(repos[j].GetPushedAt().Time)
	})

	if len(repos) == 0 {
		log.Fatalf("No repositories found for user: %v", username)
	}
	latestRepo := repos[0]

	commits, _, err := client.Repositories.ListCommits(ctx, username, *latestRepo.Name, &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		log.Fatalf("Error fetching commits: %v", err)
	}

	if len(commits) == 0 {
		log.Fatalf("No commits found for repository: %v", *latestRepo.Name)
	}

	latestCommit := commits[0]
	return latestCommit
}

func isMoreThan24HoursAgo(t time.Time) bool {
	duration := time.Since(t)
	return duration.Hours() >= 24
}
