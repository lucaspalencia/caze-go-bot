package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func printLog(text string, data interface{}) {
	empJSON, err := json.MarshalIndent(data, "", "  ")

	handleError(err, "Error in printLog()")

	fmt.Printf(""+text+"\n %s\n", string(empJSON))
}

func handleError(err error, msg string) {
	if err != nil {
		log.Println(msg)
		log.Fatal(err)
	}
}

func getClient(creds *Credentials) (*twitter.Client, error) {
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	_, _, err := client.Accounts.VerifyCredentials(verifyParams)

	if err != nil {
		return nil, err
	}

	return client, nil
}

func uploadMedia(client *twitter.Client) (*twitter.MediaUploadResult, error) {
	content, err := ioutil.ReadFile("meteuessa.jpg")

	if err != nil {
		return nil, err
	}

	media, _, err := client.Media.Upload(content, "tweet_image")

	if err != nil {
		return nil, err
	}

	printLog("Media Uploaded:", media)

	return media, nil
}

func getTweetFunc(client *twitter.Client) func(tweet *twitter.Tweet) {
	return func(tweet *twitter.Tweet) {
		printLog("Tweet to reply:", tweet)

		media, err := uploadMedia(client)

		handleError(err, "Error uploading media")

		statusParams := &twitter.StatusUpdateParams{
			InReplyToStatusID: tweet.ID,
			MediaIds:          []int64{media.MediaID},
		}

		reply, _, err := client.Statuses.Update("@"+tweet.User.ScreenName+"", statusParams)

		handleError(err, "Error replying tweet")

		printLog("Tweet replied:", reply)
	}
}

func getStream(client *twitter.Client) (*twitter.Stream, error) {
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"@caze_bot"},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := client.Streams.Filter(filterParams)

	if err != nil {
		return nil, err
	}

	return stream, nil
}

func main() {
	fmt.Println("Starting caze bot...")

	creds := Credentials{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	}

	client, err := getClient(&creds)

	handleError(err, "Error getting twitter client")

	demux := twitter.NewSwitchDemux()
	demux.Tweet = getTweetFunc(client)

	stream, err := getStream(client)

	handleError(err, "Error starting stream")

	go demux.HandleChan(stream.Messages)

	select {}

	stream.Stop()
	fmt.Println("Stopping caze bot...")
}
