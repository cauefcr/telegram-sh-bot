package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/kbinani/screenshot"
	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	admin      string
	token      string
	pollTime   int64
	debug      bool
	filePrefix = "screenshots"
)

func init() {
	flag.StringVar(&admin, "admin", "fahim_abrar", "Username of the admin")
	flag.StringVar(&token, "token", os.Getenv("TELEGRAM_BOT_TOKEN"), "Token of your bot")
	flag.Int64Var(&pollTime, "poll_time", 100, "Response time of bot")
	flag.BoolVar(&debug, "debug", false, "Print error info to debug")

	flag.Parse()
}

func main() {
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: time.Duration(pollTime) * time.Millisecond},
		Reporter: func(err error) {
			if debug {
				log.Println(err)
			}
		},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(m *tb.Message) {
		log.Println("Hi, " + m.Sender.FirstName + " " + m.Sender.LastName + "!")
		_, err := b.Send(m.Chat, "Hi, "+m.Sender.FirstName+" "+m.Sender.LastName+"!")
		if err != nil {
			log.Fatal(err)
		}
	})

	b.Handle("/sh", func(m *tb.Message) {
		log.Println(m.Sender.Username + ": " + m.Text)

		if !isAdmin(m.Sender.Username) {
			_, err := b.Send(m.Chat, "Only "+admin+" is authorized to run /sh command! You can run /hello 	\xF0\x9F\x98\x82")
			if err != nil {
				log.Println(err)
			}
			return
		}

		stringArray := strings.Split(m.Text, " ")
		if len(stringArray) == 1 {
			_, err := b.Send(m.Chat, "Please Provide Command to run!")
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		cmd := stringArray[1]
		args := make([]string, 0)

		if len(stringArray) > 2 {
			args = append(stringArray[2:])
		}

		resp, err := sh.Command(cmd, args).Output()
		if err != nil {
			log.Println(string(resp))
			_, err := b.Send(m.Chat, ""+err.Error())

			if err != nil {
				log.Fatal(err)
			}
			return
		}
		_, err = b.Send(m.Chat, ""+string(resp))
		if err != nil {
			log.Fatal(err)
		}
	})

	b.Handle("/getss", func(m *tb.Message) {
		log.Println(m.Sender.Username + ": " + m.Text)
		if !isAdmin(m.Sender.Username) {
			_, err := b.Send(m.Chat, "Only "+admin+" is authorized to run /getss command! You can run /hello 	\xF0\x9F\x98\x82")
			if err != nil {
				log.Println(err)
			}
			return
		}

		_, err := b.Send(m.Chat, "Uploading ScreenShots Please Wait, "+m.Sender.FirstName)
		if err != nil {
			log.Println(err)
		}

		fileNames, err := GetScreenShots()
		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < len(fileNames); i++ {
			ss := &tb.Photo{File: tb.FromDisk(fileNames[i])}

			_, err = b.SendAlbum(m.Sender, tb.Album{ss})
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	log.Println("Bot is running. . . . .")
	b.Start()
}

func GetScreenShots() ([]string, error) {
	n := screenshot.NumActiveDisplays()
	var fileNames = make([]string, 0)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return []string{}, err
		}

		fileNames = append(fileNames, fmt.Sprintf("%s_%d.png", filePrefix, i))
		file, _ := os.Create(fileNames[i])
		defer file.Close()
		err = png.Encode(file, img)
		if err != nil {
			return []string{}, err
		}
	}

	return fileNames, nil
}

func isAdmin(username string) bool {
	if username == admin {
		return true
	}

	return false
}
