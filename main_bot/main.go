package main

import (
    "fmt"
    "os"
    "os/signal"
    "sort"
    "strings"
    "syscall"

    "github.com/bwmarrin/discordgo"
    "github.com/go-resty/resty/v2"
    "github.com/joho/godotenv"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file")
        return
    }

    token := os.Getenv("TOKEN")
    wolframAppID := os.Getenv("WOLFRAM_APP_ID")
    guildID := "123456789012345678"
    // Replace with your actual guild ID

    dg, err := discordgo.New("Bot " + token)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return
    }

    dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
        messageCreate(s, m, wolframAppID, guildID)
    })

    err = dg.Open()
    if err != nil {
        fmt.Println("error opening connection,", err)
        return
    }

    fmt.Println("Bot is now running. Press CTRL+C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate, wolframAppID, guildID string) {
    if m.Author.ID == s.State.User.ID {
        return
    }

    if m.Content == "!ping" {
        s.ChannelMessageSend(m.ChannelID, "Pong!")
    }

    if strings.HasPrefix(m.Content, "!jawb ") {
        query := strings.TrimPrefix(m.Content, "!jawb ")
        response := queryWolframAlpha(query, wolframAppID)
        s.ChannelMessageSend(m.ChannelID, response)
    }

    if strings.HasPrefix(m.Content, "!tag ") {
        username := strings.TrimPrefix(m.Content, "!tag ")
        userID, err := getUserIDByUsername(s, guildID, username)
        if err != nil {
            s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
            return
        }
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I am a bot created by <@%s>", userID))
    }
}

func queryWolframAlpha(query, appID string) string {
    client := resty.New()
    resp, err := client.R().
        SetQueryParams(map[string]string{
            "input": query,
            "appid": appID,
        }).
        Get("http://api.wolframalpha.com/v1/result")
    if err != nil {
        return "Error querying Wolfram Alpha: " + err.Error()
    }
    return resp.String()
}

func getUserIDByUsername(s *discordgo.Session, guildID, username string) (string, error) {
    members, err := s.GuildMembers(guildID, "", 1000)
    if err != nil {
        return "", err
    }

    sort.Slice(members, func(i, j int) bool {
        return members[i].User.Username < members[j].User.Username
    })

    index := sort.Search(len(members), func(i int) bool {
        return members[i].User.Username >= username
    })

    if index < len(members) && members[index].User.Username == username {
        return members[index].User.ID, nil
    }

    return "", fmt.Errorf("user not found")
}
