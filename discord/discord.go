package discord

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

type ratingDetail struct {
	name   string
	target float64
}

const (
	lastBouns = 30
	statusLen = 3
	statusMax = 1800
	targetMax = 30000
)

type Discord interface {
	InteractCommand(s *discordgo.Session, i *discordgo.InteractionCreate)
	RegisterSlashCommand(s *discordgo.Session, i *discordgo.GuildCreate)
}

type discord struct {
	s        *discordgo.Session
	devAPPID string
}

func NewDiscordEventService(
	s *discordgo.Session,
	devAPPID string,
) Discord {
	return &discord{
		s:        s,
		devAPPID: devAPPID,
	}
}

func (dc *discord) InteractCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	if data.Name != "cal" {
		return
	}

	handleEcho(s, i, parseOptions(data.Options))
}

// RegisterSlashCommand add slash command when enter a guild
func (dc *discord) RegisterSlashCommand(s *discordgo.Session, i *discordgo.GuildCreate) {
	_, err := dc.s.ApplicationCommandBulkOverwrite(dc.devAPPID, i.ID, Commands)
	if err != nil {
		log.Fatalf("could not register commands: %s", err)
	}
	log.Printf("Create in %s, by %s", i.ID, i.OwnerID)
}

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om optionMap) {
	om = make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return
}

func getStatusTotal(status []string) (int, error) {
	if len(status) != statusLen {
		return 0, errors.New("ステータス数間違ってる")
	}

	total := 0

	for _, v := range status {
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("数字を入れてください: %s", v)
		}

		if i > statusMax {
			return 0, fmt.Errorf("ステータス上限超過: %s", v)
		}

		if i+lastBouns > statusMax {
			total += statusMax
		} else {
			total += i + lastBouns
		}

	}

	return total, nil
}

func calResultAndFormatOutput(total float64, target ...float64) string {
	result := fmt.Sprintf("ステータス総合(最終試験後): %.0f\n", total)
	const (
		base        = float64(1700) // 一位得点
		statusRitao = 2.3           // ステータス倍率
	)
	// 各ランクの評価点
	fixedTargets := []ratingDetail{
		{
			name:   "SS\t",
			target: 16000,
		},
		{
			name:   "S+\t",
			target: 14500,
		},
		{
			name:   "S \t",
			target: 13000,
		},
	}

	// 評価点の比率
	pointRatio := []struct {
		upperBound float64
		rate       float64
	}{
		{
			upperBound: float64(5000),
			rate:       float64(30),
		},
		{
			upperBound: float64(5000),
			rate:       float64(15),
		},
		{
			upperBound: float64(10000),
			rate:       float64(8),
		},
		{
			upperBound: float64(10000),
			rate:       float64(4),
		},
		{
			upperBound: float64(10000),
			rate:       float64(2),
		},
		{
			upperBound: float64(100000000),
			rate:       float64(1),
		},
	}

	if len(target) > 0 && target[0] != float64(0) {
		result += fmt.Sprintf("目標評価: %.0f\n", target[0])
		fixedTargets = []ratingDetail{
			{
				name:   "最終試験",
				target: target[0],
			},
		}
	}

	result += fmt.Sprintln("--------------------")

	for _, t := range fixedTargets {
		targetRating := t.target - base - (total)*statusRitao
		finalExamPoint := float64(0)
		for _, ptRatio := range pointRatio {
			intervalMaxRating := ptRatio.upperBound * ptRatio.rate / 100
			if targetRating > intervalMaxRating {
				targetRating -= intervalMaxRating
				finalExamPoint += ptRatio.upperBound
			} else {
				finalExamPoint += math.Ceil((targetRating / ptRatio.rate) * 100)
				break
			}
		}
		result += fmt.Sprintf("%s: %.0f\n", t.name, math.Max(finalExamPoint, 0))
	}

	return result
}

func handleEcho(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	builder := new(strings.Builder)

	targetPt := float64(0)
	if v, ok := opts["target"]; ok {
		targetPt = float64(v.IntValue())
	}

	result := strings.Split(opts["status"].StringValue(), " ")

	msg := ""

	statusTotal, err := getStatusTotal(result)

	switch {
	case targetPt > targetMax:
		builder.WriteString(fmt.Sprintf("目標上限超過: %.0f", targetPt))
	case err != nil:
		builder.WriteString(fmt.Sprintf("%s", err))
	default:
		msg = calResultAndFormatOutput(float64(statusTotal), targetPt)
		builder.WriteString(msg)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})
	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}
