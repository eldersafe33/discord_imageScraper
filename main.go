package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"net/url"
	"strings"
)

const (
	apiKey         = "custom"
	searchEngineID = "custom"
	botToken       = "custom"
)

type ImageSearchResponse struct {
	Items []struct {
		Link string `json:"link"`
	} `json:"items"`
}

func searchImage(query string) (string, error) {
	baseURL := "https://www.googleapis.com/customsearch/v1"
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("key", apiKey)
	q.Set("cx", searchEngineID)
	q.Set("searchType", "image")
	q.Set("q", query)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var searchResp ImageSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		return "", err
	}

	if len(searchResp.Items) == 0 {
		return "", fmt.Errorf("no images found for query: %s", query)
	}

	return searchResp.Items[0].Link, nil
}

func sendMessageToDiscord(session *discordgo.Session, channelID, message, imageURL string) error {
	_, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: message,
		Embed: &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: imageURL,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	discord, err := discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	discord.AddHandler(func(session *discordgo.Session, message *discordgo.MessageCreate) {
		if message.Author.Bot {
			return
		}

		if strings.HasPrefix(message.Content, "!image ") {
			query := strings.TrimPrefix(message.Content, "!image ")
			fmt.Printf("Received query: %s\n", query) // Log the received query

			imageURL, err := searchImage(query)
			if err != nil {
				fmt.Println("Error searching for image:", err)
				return
			}

			err = sendMessageToDiscord(session, message.ChannelID, "На:", imageURL)
			if err != nil {
				fmt.Println("Error sending message to Discord:", err)
				return
			}
		}
	})

	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session:", err)
		return
	}

	fmt.Println("Bot is now running.")
	<-make(chan struct{})
}
