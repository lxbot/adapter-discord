package main

import (
	"bytes"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lxbot/lxlib/v2"
	"github.com/lxbot/lxlib/v2/common"
	"github.com/lxbot/lxlib/v2/lxtypes"
)

var adapter *lxlib.Adapter
var messageCh *chan *lxtypes.Message

var client *discordgo.Session
var token string

func init() {
	adapter, messageCh = lxlib.NewAdapter()
}

func main() {
	token = os.Getenv("LXBOT_DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalln("invalid token:", "'LXBOT_DISCORD_BOT_TOKEN' にアクセストークンを設定してください")
	}
	var err error
	client, err = discordgo.New("Bot " + token)
	if err != nil {
		common.FatalLog(err)
	}
	client.AddHandler(onDiscordMessage)

	// REF: https://github.com/bwmarrin/discordgo/issues/1270
	client.Identify.Intents |= discordgo.IntentMessageContent

	err = client.Open()
	if err != nil {
		common.FatalLog("error opening connection:", err)
	}
	listenLxbotMessage()
}

func listenLxbotMessage() {
	for {
		message := <-*messageCh

		mention := ""
		if message.Mode == lxtypes.ReplyMode {
			mention = "<@" + message.User.ID + "> "
		}

		for _, v := range message.Contents {
			texts := split(v.Text, 2000-len(mention))
			for _, v := range texts {
				_, err := client.ChannelMessageSend(message.Room.ID, mention+v)
				if err != nil {
					common.DebugLog(err)
					return
				}
			}
		}
	}
}

func onDiscordMessage(sess *discordgo.Session, discord *discordgo.MessageCreate) {
	if discord.Author.ID == sess.State.User.ID {
		common.TraceLog("onDiscordMessage()", "ignore myself")
		return
	}

	text := strings.TrimSpace(discord.Content)
	if text == "" {
		common.TraceLog("onDiscordMessage()", "ignore empty message")
		return
	}

	message := &lxtypes.Message{}

	message.User = lxtypes.User{
		ID:   discord.Author.ID,
		Name: discord.Author.Username,
	}

	message.Room = lxtypes.Room{
		ID:          discord.ChannelID,
		Name:        discord.ChannelID,
		Description: "",
	}
	if st, err := client.Channel(discord.ChannelID); err == nil {
		message.Room.Name = st.Name
	}

	attachments := make([]lxtypes.Attachment, len(discord.Attachments))
	for i, v := range discord.Attachments {
		attachments[i] = lxtypes.Attachment{
			Url:         v.URL,
			Description: v.Filename,
		}
	}

	message.Contents = make([]lxtypes.Content, 0)
	message.Contents = append(message.Contents, lxtypes.Content{
		ID:          discord.ID,
		Text:        text,
		Attachments: attachments,
	})

	for _, v := range discord.Mentions {
		if v.ID == sess.State.User.ID {
			message.Mode = lxtypes.ReplyMode
		}
	}

	if raw, err := common.ToJSON(discord); err == nil {
		message.Raw = raw
	}

	adapter.Send(message)
}

func split(s string, n int) []string {
	result := make([]string, 0)
	runes := bytes.Runes([]byte(s))
	tmp := ""
	for i, r := range runes {
		tmp = tmp + string(r)
		if (i+1)%n == 0 {
			result = append(result, tmp)
			tmp = ""
		}
	}
	return append(result, tmp)
}
