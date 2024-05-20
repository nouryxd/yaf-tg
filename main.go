package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

const (
	YAF_ENDPOINT = "https://i.yaf.li/upload"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	allowedUser := os.Getenv("ALLOWED_USER")

	pref := tele.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tele.OnVideo, func(c tele.Context) error {
		if c.Sender().Username != allowedUser {
			return fmt.Errorf("Unauthorized user. User: %v", c.Sender().Username)
		}

		f := c.Message().Video
		path := fmt.Sprintf("%s", f.FileID)

		c.Send("Downloading video...")
		b.Download(&f.File, path)

		c.Send("Uploading video...")
		link, err := upload(path)
		if err != nil {
			c.Send("Something went wrong FeelsBadMan")
			return fmt.Errorf("error during upload: %v", err)
		}

		return c.Send(link)
	})

	b.Handle(tele.OnPhoto, func(c tele.Context) error {
		if c.Sender().Username != allowedUser {
			return fmt.Errorf("Unauthorized user. User: %v", c.Sender().Username)
		}

		f := c.Message().Photo
		path := fmt.Sprintf("%s", f.FileID)

		b.Download(&f.File, path)
		link, err := upload(path)
		if err != nil {
			c.Send("Something went wrong FeelsBadMan")
			return fmt.Errorf("error during upload: %v", err)
		}

		return c.Send(link)
	})

	b.Handle(tele.OnDocument, func(c tele.Context) error {
		if c.Sender().Username != allowedUser {
			return fmt.Errorf("Unauthorized user. User: %v", c.Sender().Username)
		}

		f := c.Message().Document
		path := fmt.Sprintf("%s", f.FileID)

		c.Send("Downloading...")
		b.Download(&f.File, path)

		c.Send("Uploading...")
		link, err := upload(path)
		if err != nil {
			c.Send("Something went wrong FeelsBadMan")
			return fmt.Errorf("error during upload: %v", err)
		}

		return c.Send(link)
	})

	b.Handle(tele.OnAnimation, func(c tele.Context) error {
		if c.Sender().Username != allowedUser {
			return fmt.Errorf("Unauthorized user. User: %v", c.Sender().Username)
		}

		f := c.Message().Animation
		path := fmt.Sprintf("%s", f.FileID)

		c.Send("Downloading...")
		b.Download(&f.File, path)

		c.Send("Uploading...")
		link, err := upload(path)
		if err != nil {
			c.Send("Something went wrong FeelsBadMan")
			return fmt.Errorf("error during upload: %v", err)
		}

		return c.Send(link)
	})

	b.Start()
}

func upload(path string) (string, error) {
	defer os.Remove(path)
	pr, pw := io.Pipe()
	form := multipart.NewWriter(pw)

	go func() {

		defer pw.Close()

		err := form.WriteField("name", "xd")
		if err != nil {
			os.Remove(path)
			return
		}

		file, err := os.Open(path) // path to image file
		if err != nil {
			os.Remove(path)
			return
		}

		w, err := form.CreateFormFile("file", path)
		if err != nil {
			os.Remove(path)
			return
		}

		_, err = io.Copy(w, file)
		if err != nil {
			os.Remove(path)
			return
		}

		form.Close()
	}()

	req, err := http.NewRequest(http.MethodPost, YAF_ENDPOINT, pr)
	if err != nil {
		return "Something went wrong FeelsBadMan", err
	}
	req.Header.Set("Content-Type", form.FormDataContentType())

	httpClient := http.Client{
		Timeout: 300 * time.Second,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "Something went wrong FeelsBadMan", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Something went wrong FeelsBadMan", err
	}

	var reply = string(body[:])

	return reply, nil
}
