package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"text_fetch/file_create"
	"text_fetch/text_analysis"
	"text_fetch/tg_parse"
	"text_fetch/yt_parse"
)

var (
	tg_api_id   int
	tg_api_hash string
	ctx         context.Context
)

func check(err error, msg ...string) {
	if err != nil {
		log.Printf("ПОМИЛКА:\n%v: %v\n", msg, err)
		exit()
	}
}

func main() {
	var err error
	ctx = context.Background()
	err = godotenv.Load()
	check(err, ".env")
	tg_api_hash = os.Getenv("API_APP_HASH")
	tg_api_id, err = strconv.Atoi(os.Getenv("API_APP_ID"))
	if err != nil {
		log.Fatalf("A problem with string convertation: %v\n", err)
		return
	}

	fmt.Print("Яка із платформ вас цікавить Telegram (T), Youtube (Y) чи проаналізувати вже існуючий txt файл (A)?:")
	var choice string
	_, err = fmt.Scanln(&choice)
	check(err, "choice read error")
	switch choice {
	case "T":
		msgs, chatName, err := tg_parse.TelegramParse(ctx, tg_api_id, tg_api_hash)
		check(err, "tg_parse")
		jBytes, errJ := tg_parse.MarshalJSON(msgs, chatName)
		check(errJ, "MARSHAL JSON")
		var filename string
		fmt.Print("Введіть бажану назву json та txt файла:\n")
		_, err = fmt.Scanln(&filename)
		check(err, "scan yt")
		err = file_create.JSON_parse(fmt.Sprintf("%v_tg", filename), jBytes)
		check(err, "json_parse tg")
		err = text_analysis.JsonPrepoc(fmt.Sprintf("%v_tg", filename))
		check(err, "json_preproc tg")
		err = text_analysis.TextAnalysis(fmt.Sprintf("%v_tg", filename))
		check(err, "text_analysis tg")
		exit()
		break
	case "Y":
		jsonBytes, err := yt_parse.YoutubeParse(ctx)
		check(err, "youtube parse")
		var filename string
		fmt.Print("Введіть бажану назву json та txt файла:\n")
		_, err = fmt.Scanln(&filename)
		check(err, "scan yt")
		err = file_create.JSON_parse(fmt.Sprintf("%v_yt", filename), jsonBytes)
		check(err, "json_parse yt")
		err = text_analysis.JsonPrepoc(fmt.Sprintf("%v_yt", filename))
		check(err, "json_prepoc yt")
		err = text_analysis.TextAnalysis(fmt.Sprintf("%v_yt", filename))
		check(err, "text_analysis yt")
		exit()
		break
	case "A":
		var choiceFile string
		var filename string
		fmt.Print("Хочете розпарсити json файл [j] (повинен бути в ./text_analysis/json_files папці) чи txt [t] (повинно бути ./text_analysis/txt_files папці):")
		_, err = fmt.Scanln(&choiceFile)
		check(err, "analysis choiceFile")
		switch choiceFile {
		case "j":
			fmt.Print("Назва файлу:")
			_, err = fmt.Scanln(&filename)
			check(err, "filename scan j")
			err = text_analysis.JsonPrepoc(filename)
			check(err, "jsonprepoc j analysis")
			err = text_analysis.TextAnalysis(filename)
			check(err, "text analysis, j")
			exit()
			break
		case "t":
			fmt.Print("Назва файлу:")
			_, err = fmt.Scanln(&filename)
			check(err, "filename scan t")
			err = text_analysis.TextAnalysis(filename)
			check(err, "text analysis, t")
			exit()
			break
		}
	default:
		check(errors.New("Неправильний вибір, напишіть T, Y або А для вибора між Телеграмом та Ютубом"), "mistype of type of parse")
	}
}

func exit() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Нажміть Enter для виходу з програми...")
	_, err := reader.ReadString('\n')
	check(err, "readall stdin")
	return
}
