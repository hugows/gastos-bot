package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func AllowedUsers(username string) bool {
	userList := []string{"hugows", "kfonte"}
	for _, user := range userList {
		if user == username {
			return true
		}
	}
	return false
}

type App struct {
	Bot *tgbotapi.BotAPI
	Srv *sheets.Service
	// Users    map[string]string
	Commands map[string]func(*App, *tgbotapi.Message)
}

func (app *App) AddRecord(userName string, timestamp string, spent float64, description string) error {
	sid := ""
	if sid = app.FindUserSheet(userName); sid == "" {
		return fmt.Errorf("Spreadsheet not found for user " + userName)
	}

	records := []interface{}{timestamp, spent, description}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{records},
	}

	_, err := app.Srv.Spreadsheets.Values.Append(sid, "Sheet1!A:C", valueRange).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return fmt.Errorf("error adding row: %v", err)
	}
	return nil
}

func NewApp(telegramToken, serviceKeyPath string) (*App, error) {
	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, err
	}

	// Initialize Google Sheets client
	ctx := context.Background()
	b, err := os.ReadFile(serviceKeyPath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope, sheets.DriveScope)
	if err != nil {
		return nil, err
	}
	client := config.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	commands := map[string]func(*App, *tgbotapi.Message){
		"total": func(app *App, msg *tgbotapi.Message) {
			stats, err := app.GetSpendingStats(msg.From.UserName)
			if err != nil {
				msg := tgbotapi.NewMessage(msg.Chat.ID, "Error getting total: "+err.Error())
				app.Bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(msg.Chat.ID, stats.ToString())
				app.Bot.Send(msg)
			}
		},
		"new": func(app *App, msg *tgbotapi.Message) {
			msgInitial := tgbotapi.NewMessage(msg.Chat.ID, "Creating a new Google Sheet for you...")
			app.Bot.Send(msgInitial)

			newSheet := createAndShareGoogleSheet(msg.From.UserName)
			msg2 := tgbotapi.NewMessage(msg.Chat.ID, "A new Google Sheet has been shared with you: "+newSheet.SpreadsheetUrl)
			SaveUserSheet(msg.From.UserName, newSheet.SpreadsheetId)
			app.Bot.Send(msg2)
		},
	}

	return &App{Bot: bot, Srv: srv, Commands: commands}, nil
}

func main() {

	TELEGRAM_TOKEN := os.Getenv("TELEGRAM_TOKEN")
	if TELEGRAM_TOKEN == "" {
		log.Panic("TELEGRAM_TOKEN not found")
	}

	app, err := NewApp(TELEGRAM_TOKEN, "service-key.json")
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := app.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if app.FindUserSheet(update.Message.From.UserName) == "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Spreadsheet not found for user "+update.Message.From.UserName)
				app.Bot.Send(msg)
				app.Commands["new"](app, update.Message)
			} else {
				spent, err := ExtractPriceAndDescription(update.Message.Text)
				if err != nil {
					// Check if a command and run
					if cmd, ok := app.Commands[strings.TrimSpace(strings.ToLower(update.Message.Text))]; ok {
						cmd(app, update.Message)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Not a command / invalid value: "+update.Message.Text)
						app.Bot.Send(msg)
					}
				} else if spent.Price <= 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Smaller than zero makes no sense here!")
					app.Bot.Send(msg)
				} else {
					timestamp := update.Message.Time().Format("2006-01-02 15:04:05")
					if err := app.AddRecord(update.Message.From.UserName, timestamp, spent.Price, spent.DescriptionWithFallback()); err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error adding record: "+err.Error())
						app.Bot.Send(msg)
					} else {
						app.Commands["total"](app, update.Message)
					}
					// time.Sleep(3 * time.Second)
					// msgDelete := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, msgSent.MessageID)
					// bot.Send(msgDelete)
				}

			}

		}
	}
}

func createAndShareGoogleSheet(userName string) *sheets.Spreadsheet {
	ctx := context.Background()
	b, err := os.ReadFile("service-key.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope, sheets.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := config.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Create a new spreadsheet and share with the user
	sheet, err := srv.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: userName + "'s - Gastos Mensais",
		},
	}).Do()
	if err != nil {
		log.Fatalf("Unable to create a new spreadsheet: %v", err)
	}

	// *** Create the permission ***
	permissions := &drive.Permission{
		Type: "anyone",
		Role: "writer",
		// AllowFileDiscovery: false, // Prevent the file from appearing in shared searches
	}

	driveSrv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	// // *** Call the Drive API to apply the permission ***
	_, err = driveSrv.Permissions.Create(sheet.SpreadsheetId, permissions).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Unable to set share permissions: %v", err)
	}

	// Add a header row
	records := []interface{}{"Data", "Valor", "Descrição", "Categoria"}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{records},
	}

	_, err3 := srv.Spreadsheets.Values.Append(sheet.SpreadsheetId, "Sheet1!A:D", valueRange).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err3 != nil {
		fmt.Println("Error adding header:", err)
	}

	return sheet
}
