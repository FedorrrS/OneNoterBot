package telegram

import (
	"OneNoterBot/internal/response"
	"OneNoterBot/pkg/e"
	"OneNoterBot/pkg/logging"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/emirpasic/gods/maps/linkedhashmap"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
)

type Bot struct {
	bot *tgbotapi.BotAPI
}

func NewBot(bot *tgbotapi.BotAPI) *Bot {
	return &Bot{bot: bot}
}

func (b *Bot) Start(logger *logging.Logger, db *sql.DB) {
	updConfig := tgbotapi.NewUpdate(0)
	updConfig.Timeout = 30
	upds := b.bot.GetUpdatesChan(updConfig)

	var (
		password        string
		curCommand      = make(map[string]bool, 2)
		numberedNotes   = linkedhashmap.New()
		verificationMsg string
	)

	for upd := range upds {
		if upd.Message == nil {
			continue
		}

		lang := b.Lang(upd, logger)
		firstname := upd.Message.From.FirstName
		resp := response.NewResponse(lang)
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "")

		if upd.Message.IsCommand() {
			switch upd.Message.Command() {
			case start:
				curCommand[start] = true
				msg.Text = resp.Greeting()
			case help:
				msg.Text = resp.Help()
			case notes:
				if password == "" {
					msg.Text = resp.AuthorizationFailed()
					msg.ReplyToMessageID = upd.Message.MessageID
				} else {
					getNotes(numberedNotes, db, password, logger)
				}
				if numberedNotes.Size() == 0 {
					msg.Text = resp.EmptyNotes()
				} else {
					msg.Text = resp.GiveNotes(numberedNotes.Values())
				}
			case clear:
				getNotes(numberedNotes, db, password, logger)
				msg.Text = resp.ClearVerification()
				curCommand[clear] = ClearHandler(password, &msg, resp, upd, numberedNotes)
			case clearall:
				getNotes(numberedNotes, db, password, logger)
				msg.Text = resp.ClearallVerification()
				curCommand[clearall] = ClearHandler(password, &msg, resp, upd, numberedNotes)
			case whoami:
				WhoamiHandler(password, &msg, resp, upd)
			default:
				msg.Text = resp.CommandNotSupported()
				msg.ReplyToMessageID = upd.Message.MessageID
			}
		} else {
			switch {
			case curCommand[start]:
				password = upd.Message.Text
				curCommand[start] = false
				msg.Text = resp.AuthorizationSuccess(firstname)
			case curCommand[clear]:
				curCommand[clear] = false
				splitMsg := lo.Map(strings.Split(upd.Message.Text, ","), func(s string, _ int) string {
					return strings.TrimSpace(s)
				})
				verificationMsg = strings.ToLower(splitMsg[0])
				switch {
				case resp.IsPositive(verificationMsg):
					if len(splitMsg) == 1 {
						msg.Text = resp.ClearIncorrect()
					} else {
						key := splitMsg[1]
						v, _ := strconv.Atoi(key)
						if v > numberedNotes.Size() {
							msg.Text = resp.ClearNo()
							b.SendMessage(logger, msg)
							continue
						}
						el, ok := numberedNotes.Get(v)
						if !ok {
							logger.Fatal("Error: cannot get value from notes map")
						}
						if _, err := db.Exec(fmt.Sprintf("DELETE FROM `users` WHERE `password` = '%s' AND `note` = '%s'", password, strings.TrimPrefix(el.(string), fmt.Sprintf("%d. ", v)))); err != nil {
							logger.Fatal(e.Wrap("Error: deleting all notes from db", err))
						}
						msg.Text = resp.ClearYes()
					}
				case resp.IsNegative(verificationMsg):
					msg.Text = resp.ClearallNo()
				default:
					msg.Text = resp.ClearallIncorrect()
				}
			case curCommand[clearall]:
				verificationMsg = strings.TrimSpace(strings.ToLower(upd.Message.Text))
				switch {
				case resp.IsPositive(verificationMsg):
					curCommand[clearall] = false
					if _, err := db.Exec(fmt.Sprintf("DELETE FROM `users` WHERE `password` = '%s'", password)); err != nil {
						logger.Fatal(e.Wrap("Error: deleting single note from db", err))
					}
					numberedNotes.Clear()
					msg.Text = resp.ClearallYes()
				case resp.IsNegative(verificationMsg):
					curCommand[clearall] = false
					msg.Text = resp.ClearallNo()
				default:
					msg.Text = resp.ClearallIncorrect()
				}
			default:
				notes := strings.Split(upd.Message.Text, ",")
				for _, note := range notes {
					if _, err := db.Exec(fmt.Sprintf("INSERT INTO users (password, note) VALUES('%s', '%s')", password, strings.TrimSpace(note))); err != nil {
						logger.Fatal(e.Wrap("Error: inserting data to db", err))
					}
				}
				msg.Text = resp.DataSavedSuccess(firstname)
				msg.ReplyToMessageID = upd.Message.MessageID
			}
		}
		b.SendMessage(logger, msg)
	}
}

func getNotes(numberedNotes *linkedhashmap.Map, db *sql.DB, password string, logger *logging.Logger) {
	numberedNotes.Clear()
	smlp, err := db.Query(fmt.Sprintf("SELECT note FROM users WHERE password = '%s'", password))
	if err != nil {
		logger.Fatal(e.Wrap("Error: selecting notes in db", err))
	}
	i := 1
	for smlp.Next() {
		var note string
		if err := smlp.Scan(&note); err != nil {
			logger.Fatal(e.Wrap("Error: scanning sample note", err))
		}
		numberedNotes.Put(i, fmt.Sprintf("%d. %s", i, note))
		i++
	}
}

func (b *Bot) Lang(upd tgbotapi.Update, logger *logging.Logger) string {
	lang := upd.Message.From.LanguageCode
	switch lang {
	case "en", "ru":
	default:
		logger.Infof("%v language is not supported", lang)
		lang = "en"
	}
	return lang
}

func (b *Bot) SendMessage(logger *logging.Logger, msg tgbotapi.MessageConfig) {
	if _, err := b.bot.Send(msg); err != nil {
		logger.Fatal(e.Wrap("Error: sending message to user", err))
	}
}
