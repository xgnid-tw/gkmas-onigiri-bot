package discord

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "cal",
		Description: "学マス最終試験計算機",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "status",
				Description: "例： 700 1120 1400",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "target",
				Description: "目標評価点があれば、target を設定してください",
				Type:        discordgo.ApplicationCommandOptionInteger,
			},
		},
	},
}
