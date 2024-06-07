package yt_parse

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Comment struct {
	ID    string `json:"id"`
	Date  string `json:"date"`
	Actor string `json:"actor"`
	Text  string `json:"text"`
}
type MessagesData struct {
	Title    string    `json:"title"`
	VideoURL string    `json:"video_url"`
	Comments []Comment `json:"messages"`
}

var o bool

func YoutubeParse(ctx context.Context) ([]byte, error) {
	var videoId string
	o = false
	absPath, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	credDir := filepath.Join(absPath, ".credentials")
	b, err := os.ReadFile(filepath.Join(credDir, "ytclient_secret.json"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/youtube-go-quickstart.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope, youtube.YoutubeScope, youtube.YoutubeUploadScope, youtube.YoutubeForceSslScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client, err := getClient(ctx, config)
	if err != nil {
		return nil, err
	}
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	if o {
		rd := bufio.NewReader(os.Stdin)
		_, err = rd.ReadString('\n')
		if err != nil {
			return nil, err
		}
	}
	fmt.Print("Введіть url для відео ютуб коментарії з якого хочете взяти:")
	_, err = fmt.Scanln(&videoId)
	if err != nil {
		return nil, err
	}
	videoId, err = video_url_Parse(videoId)
	if err != nil {
		return nil, err
	}
	r, err := commentsListByID(service, []string{"snippet"}, videoId)
	if err != nil {
		return nil, err
	}
	jsonFile, err := MarshalJSON(r)
	if err != nil {
		return nil, err
	}
	return jsonFile, nil
}

func MarshalJSON(response *youtube.CommentThreadListResponse) ([]byte, error) {
	// Convert the input messages to our defined Message struct
	var convertedComments []Comment
	for _, m := range response.Items {
		convertedComments = append(convertedComments, Comment{
			ID:    fmt.Sprintf("%v", m.Snippet.TopLevelComment.Id),
			Date:  fmt.Sprintf("%v", m.Snippet.TopLevelComment.Snippet.PublishedAt),
			Actor: fmt.Sprintf("%v", m.Snippet.TopLevelComment.Snippet.AuthorDisplayName),
			Text:  fmt.Sprintf("%v", m.Snippet.TopLevelComment.Snippet.TextOriginal),
		})
	}

	// Wrap the messages with the outer structure
	data := MessagesData{
		Title:    "Youtube",
		VideoURL: fmt.Sprintf("https://www.youtube.com/watch?v=%v", response.Items[0].Snippet.VideoId),
		Comments: convertedComments,
	}

	// Marshal the data to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil

}

func token_url_Parse(url string) (string, error) {
	var code string

	n, found := strings.CutPrefix(url, "http://localhost/?state=state-token&code=")
	if !found {
		return "", errors.New("префікс URL для токена не знайдений, неможливий парсинг URL")
	}
	code, _, found = strings.Cut(n, "&scope=")
	return code, nil
}

func video_url_Parse(url string) (string, error) {
	id, found := strings.CutPrefix(url, "https://www.youtube.com/watch?v=")
	if !found {
		return "", errors.New("Префікс неможливо було запарсити, так як його не знайдено. Надішліть правильний url")
	}
	return id, nil
}

func commentsListByID(service *youtube.Service, part []string, videoID string) (*youtube.CommentThreadListResponse, error) {
	call := service.CommentThreads.List(part)
	response, err := call.MaxResults(1000).VideoId(videoID).Do()
	if err != nil {
		return nil, err
	}
	return response, nil
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}
		err = saveToken(cacheFile, tok)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(ctx, tok), nil
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	o = true
	var err error
	tok := new(oauth2.Token)
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Для роботи з Ютубом: \n"+
		"1) Пройдіть по посиланню та авторизуйтесь: \n%v\n"+
		"2) Cкопіюйте сюди останнє посилання після логіну в гугл аккаунт \n"+
		"(формат має бути: http://localhost/?state=state-token&code=[code]&scope=[scope_urls]):\n", authURL)
	var code string

	if _, err := fmt.Scan(&code); err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to read authorization code %v", err))
	}

	code, err = token_url_Parse(code)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("code couldn't be parsed: %v\n", err))
	}
	tok, err = config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to retrieve token from web %v", err))
	}
	return tok, nil
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	absPath, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(absPath, ".credentials")
	err = os.MkdirAll(tokenCacheDir, 0700)
	if err != nil {
		return "", err
	}
	return filepath.Join(tokenCacheDir, url.QueryEscape("user_cred.json")), nil
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	if err != nil {

	}
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return err
	}
	return nil
}
