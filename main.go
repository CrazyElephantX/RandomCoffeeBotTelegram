package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	db  *sql.DB
	bot *tgbotapi.BotAPI
)

func initDB() *sql.DB {
	// Загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Формируем строку подключения
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Создаем таблицу, если её нет
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL UNIQUE,
			username TEXT,
			first_name TEXT,
			last_name TEXT,
			registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func handleMatchCommand(chatID int64) {
	rows, err := db.Query("SELECT user_id FROM users")
	if err != nil {
		log.Println("DB query error:", err)
		return
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		users = append(users, userID)
	}

	if len(users) < 2 {
		msg := tgbotapi.NewMessage(chatID, "Недостаточно участников для мэтчинга.")
		bot.Send(msg)
		return
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(users), func(i, j int) {
		users[i], users[j] = users[j], users[i]
	})

	for i := 0; i < len(users); i += 2 {
		if i+1 >= len(users) {
			// Нечетное количество - последний остается без пары
			msg := tgbotapi.NewMessage(users[i], "К сожалению, не нашлось пары для вас в этом раунде.")
			bot.Send(msg)
			break
		}

		user1 := users[i]
		user2 := users[i+1]

		msg1 := tgbotapi.NewMessage(user1, fmt.Sprintf("Ваш партнёр для Random Coffee: @%s", getUserUsername(user2)))
		msg2 := tgbotapi.NewMessage(user2, fmt.Sprintf("Ваш партнёр для Random Coffee: @%s", getUserUsername(user1)))

		bot.Send(msg1)
		bot.Send(msg2)
	}
}

func initBot() *tgbotapi.BotAPI {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized as %s", bot.Self.UserName)
	return bot
}

func main() {
	log.Println("Starting bot...")

	// Инициализация
	db = initDB()
	defer db.Close()
	bot = initBot()

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				var exists bool
				err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", update.Message.From.ID).Scan(&exists)
				if err != nil {
					log.Println("DB error:", err)
					continue
				}

				if exists {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы уже зарегистрированы!")
					bot.Send(msg)
				} else {
					_, err := db.Exec(
						"INSERT INTO users (user_id, username, first_name, last_name) VALUES ($1, $2, $3, $4)",
						update.Message.From.ID,
						update.Message.From.UserName,
						update.Message.From.FirstName,
						update.Message.From.LastName,
					)
					if err != nil {
						log.Println("Insert error:", err)
						continue
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы успешно зарегистрированы в Random Coffee!")
					bot.Send(msg)
				}
			case "match":
				handleMatchCommand(update.Message.Chat.ID)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда")
				bot.Send(msg)
			}
		}
	}
}

func getUserUsername(userID int64) string {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE user_id = $1", userID).Scan(&username)
	if err == sql.ErrNoRows {
		return "пользователь"
	} else if err != nil {
		log.Println("Error getting username:", err)
		return "пользователь"
	}
	return username
}
