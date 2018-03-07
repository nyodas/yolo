package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"github.com/coreos/etcd/client"
)

func listBucket() (remoteCmdSound []*cobra.Command) {
	bucketName := "yolo-sound"
	ctx := context.Background()
	query := &storage.Query{}
	client, err := storage.NewClient(ctx,option.WithoutAuthentication())
	if err != nil {
		log.Fatal(err)
	}

	it := client.Bucket(bucketName).Objects(ctx,query)
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("listBucket: unable to list bucket %q: %v", bucketName, err)
			return
		}
		if obj.Name == "favicon.ico" {
			continue
		}
		sndName := strings.TrimSuffix(obj.Name,".mp3")
		sndCmd := cobra.Command{
			Use:   sndName + "",
			Short: "Play " + sndName,
			Run: func(cmd *cobra.Command, args []string) {
				sound := snd{}
				sound.SmartUrl(sndName)
				sound.Play()
			},
		}
		remoteCmdSound = append(remoteCmdSound, &sndCmd)
	}
	return
}

type snd struct {
	File string `json:"file"`
}

func (s *snd) toJson() (string){
	jsonByte,err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(jsonByte)
}

func (s *snd) SmartUrl(sndName string) {
	sndUrl ,_:= url.Parse(string("https://storage.googleapis.com/yolo-sound/"+ sndName + ".mp3"))
	s.File = sndUrl.String()
}

func (s *snd) Play(){
	cfg := client.Config{
		Endpoints:               []string{"https://etcd.snd.wtf"},
		Transport:               client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)
	_, err = kapi.Set(context.Background(), "/yolo_grafana", s.toJson(), nil)
	if err != nil {
		log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Set is done, it should begin soon")
	}
}



func main() {
	sndSoundCmds := listBucket()
	var cmdPrint = &cobra.Command{
		Use:   "url [url to play]",
		Short: "Play a sound url on yolo",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sound := snd{
				File: args[0],
			}
			sound.Play()
		},
	}

	var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.AddCommand(cmdPrint)
	for _,cmd := range sndSoundCmds {
		rootCmd.AddCommand(cmd)
	}
	rootCmd.Execute()
}