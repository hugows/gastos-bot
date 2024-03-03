# Gastos bot

Your bot will confirm the entry and the expense will be logged in the specified Google Sheets document.

## Custom Commands

- `/start` - Initialize interaction with the bot.
- `/total` - Receive a summary of the spending in the current Sheet.

## How to run

1. Create a Google Sheets project and download a service credentials JSON file. Save as "service-key.json" in the root of the repo.
2. Create a new Telegram Bot (talking to @BotFather) and take note of the token
3. TELEGRAM_TOKEN=1234 go run .
