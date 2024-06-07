package tg_parse

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	teleg "github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Message struct {
	ID      int    `json:"id"`
	Date    string `json:"date"`
	Actor   string `json:"actor"`
	ReplyId string `json:"reply_to_message_id"`
	Text    string `json:"text"`
}

type MessagesData struct {
	Title    string    `json:"title"`
	ChatName string    `json:"chat_name"`
	Messages []Message `json:"messages"`
}

var (
	phone         string
	chat_username string
	password      string
)

func TelegramParse(ctx context.Context, api_id int, api_hash string) ([]*tg.Message, string, error) {
	Messages := make([]*tg.Message, 0)

	api_hash = os.Getenv("API_APP_HASH")
	client := teleg.NewClient(api_id, api_hash, teleg.Options{})
	if err := client.Run(ctx, func(ctx context.Context) error {
		defer func() {
			if _, err := client.API().AuthLogOut(ctx); err != nil {
				log.Printf("Failed to log out: %v", err)
			} else {
				fmt.Println("Logged out successfully")
			}
		}()
		//checks if the password needed. DO NOT CHANGE errCodeVAR NAME
		_, errCode := auth.CodeOnly(phone, auth.CodeAuthenticatorFunc(codeAsk)).Password(ctx)
		tokMap, err := createAuthToken(errCode)
		if err != nil {
			return errors.New(fmt.Sprintf("помилка з auth файлом: %v\n", err))
		}
		if tokMap != nil {
			phone = tokMap["phone"]
			password = tokMap["password"]
		}
		fmt.Println(phone)
		if errors.Is(errCode, auth.ErrPasswordNotProvided) {
			if err != nil {
				return err
			}
			err = auth.NewFlow(
				auth.Constant(phone, password, auth.CodeAuthenticatorFunc(codeAsk)),
				auth.SendCodeOptions{},
			).Run(ctx, client.Auth())
			if err != nil {
				return err
			}
		} else {
			err = auth.NewFlow(
				auth.CodeOnly(phone, auth.CodeAuthenticatorFunc(codeAsk)),
				auth.SendCodeOptions{},
			).Run(ctx, client.Auth())
			if err != nil {
				return err
			}
		}
		fmt.Print("Введіть Username чату з якого хочете взяти повідомлення:")
		_, err = fmt.Scanln(&chat_username)
		Messages, err = MessageFetch(ctx, client.API(), chat_username)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, "", err
	}

	return Messages, chat_username, nil
}

func codeAsk(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Введіть код прийшовший на телеграм з вище наведеним номером телефону для авторизації:")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println("Problem with reading a string")
		return "", err
	}
	code = strings.ReplaceAll(code, "\n", "")
	return code, nil
}
func MessageFetch(ctx context.Context, client *tg.Client, username string) ([]*tg.Message, error) {
	// Search for a public chat by username
	chat, err := client.ContactsResolveUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	channel := chat.Chats[0].(*tg.Channel)

	var limit int

	fmt.Print("Ліміт повідомлень? (Максимум 2700):")
	_, err = fmt.Scanln(&limit)
	if limit > 2700 {
		return nil, errors.New("Ліміт перевищує 2700, неможливо передати повідомлення")
	}
	messages := make([]*tg.Message, 0)
	var offsetID int
	fmt.Println("Фетчинг повідомлень...")
	for len(messages) < limit {
		fetchLimit := 100
		if limit-len(messages) < 100 {
			fetchLimit = limit - len(messages)
		}

		history, err := client.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer: &tg.InputPeerChannel{
				ChannelID:  channel.ID,
				AccessHash: channel.AccessHash,
			},
			AddOffset:  0,
			Limit:      fetchLimit,
			MaxID:      0,
			MinID:      0,
			OffsetID:   offsetID,
			OffsetDate: 0,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get message history: %w", err)
		}

		messageClasses, ok := history.(*tg.MessagesChannelMessages)
		if !ok {
			return nil, fmt.Errorf("unexpected type for message history")
		}

		for _, msg := range messageClasses.Messages {
			if message, ok := msg.(*tg.Message); ok {
				messages = append(messages, message)
			}
		}

		if len(messageClasses.Messages) == 0 {
			break // No more messages to fetch
		}

		// Set the offset ID to the ID of the last message fetched
		offsetID = messageClasses.Messages[len(messageClasses.Messages)-1].(*tg.Message).ID
	}
	return messages, nil
}
func createAuthToken(e error) (map[string]string, error) {
	// Checking if the file exists, if not, creating a new file with data in it
	absPath, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(absPath, ".credentials", "auth_token.json")
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("помилка з відкриттям файлу для авторизація: %v\n", err))
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("помилка прочитання файла: %v\n", err))
	}

	// If the file is empty, prompt the user for credentials
	if len(b) == 0 {
		fmt.Print("Введіть свій номер телефону для входу в телеграм:")
		_, err = fmt.Scanln(&phone)
		if err != nil {
			return nil, err
		}

		if errors.Is(e, auth.ErrPasswordNotProvided) {
			fmt.Print("Введіть пароль (тільки якщо у вас включена 2-factor auth) або введіть na (якщо 2-factor auth виключена):")
			_, err = fmt.Scanln(&password)
			if err != nil {
				return nil, err
			}
		}

		// Truncate the file before writing new content
		err = f.Truncate(0)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("помилка з truncate файлу для авторизації: %v\n", err))
		}

		// Move the file pointer to the beginning before writing
		_, err = f.Seek(0, 0)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("помилка з seek файлу для авторизації: %v\n", err))
		}

		_, err := f.Write([]byte(fmt.Sprintf(`{"phone":"%v","password":"%v"}`, phone, password)))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("помилка з написанням файлу для авторизації: %v\n", err))
		}

		// Move the file pointer to the beginning before reading the new content
		_, err = f.Seek(0, 0)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("помилка з seek файлу для прочитання: %v\n", err))
		}

		b, err = io.ReadAll(f)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Помилка з прочитанням написаного файла"))
		}
	}

	authCred := make(map[string]string)
	err = json.Unmarshal(b, &authCred)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("помилка з парсингом json файлу: %v\n", err))
	}
	return authCred, nil
}
func MarshalJSON(messages []*tg.Message, chatName string) ([]byte, error) {
	// Convert the input messages to our defined Message struct
	var convertedMessages []Message
	for _, m := range messages {
		postAuthor, _ := m.GetPostAuthor()
		convertedMessages = append(convertedMessages, Message{
			ID:    m.GetID(),
			Date:  fmt.Sprintf("%v", m.GetDate),
			Actor: fmt.Sprintf("%v", postAuthor),
			Text:  m.GetMessage(),
		})
	}

	// Wrap the messages with the outer structure
	data := MessagesData{
		Title:    "Telegram",
		ChatName: fmt.Sprintf("%v", chatName),
		Messages: convertedMessages,
	}

	// Marshal the data to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}
