package main

import (
	"bytes"
	"encoding/gob"
	"github.com/bwmarrin/discordgo"
	"github.com/lxbot/lxlib"
	"log"
	"os"
	"strings"
)

type M = map[string]interface{}

var ch *chan M
var client *discordgo.Session
var token string

func Boot(c *chan M) {
	ch = c

	gob.Register(discordgo.MessageCreate{})
	gob.Register(discordgo.Message{})
	gob.Register(discordgo.MessageAttachment{})
	gob.Register(discordgo.MessageEmbed{})
	gob.Register(discordgo.MessageEmbedAuthor{})
	gob.Register(discordgo.MessageEmbedField{})
	gob.Register(discordgo.MessageEmbedFooter{})
	gob.Register(discordgo.MessageEmbedImage{})
	gob.Register(discordgo.MessageEmbedProvider{})
	gob.Register(discordgo.MessageEmbedThumbnail{})
	gob.Register(discordgo.MessageEmbedVideo{})
	gob.Register(discordgo.MessageReaction{})
	gob.Register(discordgo.Emoji{})
	gob.Register(discordgo.Member{})

	token = os.Getenv("LXBOT_DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalln("invalid token:", "'LXBOT_DISCORD_BOT_TOKEN' にアクセストークンを設定してください")
	}
	var err error
	client, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln(err)
	}
	client.AddHandler(onMessage)

	err = client.Open()
	if err != nil {
		log.Println("error opening connection:", err)
	}
}

func Send(msg M) {
	m, err := lxlib.NewLXMessage(msg)
	if err != nil {
		log.Println(err)
		return
	}

	texts := split(m.Message.Text, 2000)
	for _, v := range texts {
		_, err := client.ChannelMessageSend(m.Room.ID, v)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func Reply(msg M) {
	m, err := lxlib.NewLXMessage(msg)
	if err != nil {
		log.Println(err)
		return
	}

	mention := "<@" + m.User.ID + "> "

	texts := split(m.Message.Text, 2000 - len(mention))
	for _, v := range texts {
		_, err := client.ChannelMessageSend(m.Room.ID, mention + v)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func onMessage(sess *discordgo.Session, msg *discordgo.MessageCreate) {
	if msg.Author.ID == sess.State.User.ID {
		return
	}

	text := strings.TrimSpace(msg.Content)

	isReply := false
	for _, v := range msg.Mentions {
		if v.ID == sess.State.User.ID {
			isReply = true
		}
	}

	attachments := make([]M, len(msg.Attachments))
	for i, v := range msg.Attachments {
		attachments[i] = M{
			"url": v.URL,
			"description": v.Filename,
		}
	}

	channelName := msg.ChannelID
	topic := ""
	if i, err := client.Channel(msg.ChannelID); err == nil {
		channelName = i.Name
		topic = i.Topic
	}

	*ch <- M{
		"user": M{
			"id":   msg.Author.ID,
			"name": msg.Author.Username,
		},
		"room": M{
			"id":          msg.ChannelID,
			"name":        channelName,
			"description": topic,
		},
		"message": M{
			"id":          msg.ID,
			"text":        text,
			"attachments": attachments,
		},
		"is_reply":  isReply,
		"raw":       *msg,
	}
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
