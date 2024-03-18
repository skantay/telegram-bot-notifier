package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/skantay/notifier/config"
	"github.com/skantay/notifier/internal/bot"
	"github.com/skantay/notifier/internal/bot/middleware"
	"github.com/skantay/notifier/internal/botkit"
	"github.com/skantay/notifier/internal/fetcher"
	"github.com/skantay/notifier/internal/notifier"
	"github.com/skantay/notifier/internal/repository"
	"github.com/skantay/notifier/internal/summary"
)

func main() {
	cfg := config.Get()
	botAPI, err := tgbotapi.NewBotAPI(cfg.OpenAIKey)
	if err != nil {
		log.Printf("tgbotapi error: %v", err)

		return
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseDSN)
	if err != nil {
		log.Printf("db error: %v", err)

		return
	}
	defer db.Close()

	articleRepo := repository.NewArticlerepository(db)
	sourceRepo := repository.NewSourcerepository(db)

	fetcher := fetcher.New(articleRepo, sourceRepo, cfg.FetchInterval, cfg.FilterKeywords)

	notifier := notifier.New(articleRepo, summary.NewOpenAISummarizer(cfg.OpenAIKey, cfg.OpenAIPrompt), botAPI, cfg.NotificationInterval, 2*cfg.FetchInterval, cfg.TelegramChannelID)

	ctx, cancle := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancle()

	newsBot := botkit.New(botAPI)

	newsBot.RegisterCMDView("start", bot.ViewCmdStart())

	newsBot.RegisterCMDView(
		"addsource",
		middleware.AdminsOnly(cfg.TelegramChannelID, bot.ViewCmdAddSource(sourceRepo)),
	)

	go func(ctx context.Context) {
		if err := fetcher.Start(ctx); err != nil {
			log.Printf("fetcher error: %v", err)

			return
		}
	}(ctx)

	go func(ctx context.Context) {
		if err := notifier.Start(ctx); err != nil {
			log.Printf("notifier error: %v", err)

			return
		}
	}(ctx)

	if err := newsBot.Run(ctx); err != nil {
		return
	}
}
