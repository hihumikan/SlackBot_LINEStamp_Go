package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Autogenerated struct {
	Packageid           int        `json:"packageId"`
	Onsale              bool       `json:"onSale"`
	Validdays           int        `json:"validDays"`
	Title               Title      `json:"title"`
	Author              Author     `json:"author"`
	Price               []Price    `json:"price"`
	Stickers            []Stickers `json:"stickers"`
	Hasanimation        bool       `json:"hasAnimation"`
	Hassound            bool       `json:"hasSound"`
	Stickerresourcetype string     `json:"stickerResourceType"`
}
type Title struct {
	En   string `json:"en"`
	ZhTw string `json:"zh_TW"`
}
type Author struct {
	En   string `json:"en"`
	ZhTw string `json:"zh_TW"`
}
type Price struct {
	Country  string  `json:"country"`
	Currency string  `json:"currency"`
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
}
type Stickers struct {
	ID     int `json:"id"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
type Ping struct {
	Status int
	Rssult string
}

func main() {
	//Firebase Admin SDK
	ctx := context.Background()
	sa := option.WithCredentialsFile("slackbot-app-mimimi-firebase-adminsdk-xwhaq-09a3dd413b.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	/*

		//データの追加
		_, _, err = client.Collection("users").Add(ctx, map[string]interface{}{
			"first": "Ada",
			"last":  "Lovelace",
			"born":  1815,
		})

	*/
	//データを読み取り
	iter := client.Collection("stampID").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		fmt.Println(doc.Data())
	}

	length := 100
	flag := 0
	capacity := 100
	array := make([]int, length, capacity)

	// .envの読み取り
	godotenv.Load(fmt.Sprintf("./%s.env", os.Getenv("GO_ENV")))

	// SlackClientの構築
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	// ルートにアクセスがあった時の処理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ping := Ping{http.StatusOK, "ok"}

		res, err := json.Marshal(ping)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		switch eventsAPIEvent.Type {
		case slackevents.URLVerification: // URL検証の場合の処理
			var res *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(res.Challenge)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case slackevents.CallbackEvent: // コールバックイベントの場合の処理
			innerEvent := eventsAPIEvent.InnerEvent

			attachment := slack.Attachment{
				Pretext: "【使い方】LINE STAMP BOTのコマンド一覧",
				Text:    "list",
				// Uncomment the following part to send a field too

				Fields: []slack.AttachmentField{
					slack.AttachmentField{
						Title: "`?s random`",
						Value: "LINEスタンプがランダム表示されます",
					},
					slack.AttachmentField{
						Title: "`?s add '任意のスタンプID`",
						Value: "スタンプに割り当てられたIDを登録します",
					},
					slack.AttachmentField{
						Title: "`?s show インデックス番号`",
						Value: "スタンプに割り当てられたIDのURLを返します",
					},
					slack.AttachmentField{
						Title: "`?s urlid 任意のスタンプショップID`",
						Value: "スタンプショップのIDに登録されたLINEスタンプを登録します",
					},
					slack.AttachmentField{
						Title: "`?s help`",
						Value: "helpを表示します",
					},
				},
			}

			message := slack.MsgOptionAttachments(attachment)

			// イベントタイプで分岐
			switch event := innerEvent.Data.(type) {
			case *slackevents.MessageEvent: // メッセージイベント
				fmt.Println("メッセージの受信")
				if strings.Index(event.Text, "?s random") != -1 {
					// 送信元のユーザIDを取得
					//user := event.User

					rand.Seed(time.Now().UnixNano())

					var i int = rand.Intn(96666666)
					var s string

					s = strconv.Itoa(i)
					var byte_buf bytes.Buffer
					byte_buf.WriteString("https://stickershop.line-scdn.net/stickershop/v1/sticker/")
					byte_buf.WriteString(s)
					byte_buf.WriteString("/iPhone/sticker@2x.png")

					// 送信元ユーザに注意
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(byte_buf.String(), false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else if strings.Index(event.Text, "?s add") == 0 {

					//fmt.Println(event.Text)

					tmp := strings.Split(event.Text, " ")

					// デバッグ用
					//fmt.Println(event)

					if len(tmp) < 3 {
						if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("?s add \"ここに任意の文字列を入力してください\" ", false)); err != nil {
							log.Println(err)
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
					}

					for i := flag; i < flag+1; i++ {
						array[i], _ = strconv.Atoi(tmp[2])
					}
					flag++
					//fmt.Println(flag)
					//fmt.Println(array[flag-1])

					for i := 0; i < len(array); i++ {
						if array[i] != 0 {
							//fmt.Println(array[i])
						}
					}

					firebaseNum, _ := strconv.Atoi(tmp[2])
					//fmt.Println(firebaseNum)
					_, _, err = client.Collection("stampID").Add(ctx, map[string]interface{}{
						"ID": firebaseNum,
					})
					// iter := client.Collection("stampID").Where("ID", "==", firebaseNum).Documents(ctx)
					iter := client.Collection("stampID").Documents(ctx)
					if err != nil {
						return
					}
					for {
						doc, err := iter.Next()
						if err == iterator.Done {
							break
						}
						if err != nil {
							return
						}
						fmt.Println(doc.Data())
					}

					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("追加しました ", false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					var byte_buf bytes.Buffer
					s := strconv.Itoa(flag - 1)
					byte_buf.WriteString("\" ?s show ")
					byte_buf.WriteString(s)
					byte_buf.WriteString("\"で表示できます")
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(byte_buf.String(), false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					//flag = event.Text
				} else if strings.Index(event.Text, "?s show") == 0 {
					//fmt.Println(event.Text)
					//

					tmp := strings.Split(event.Text, " ")

					// デバッグ用
					//fmt.Println(event)

					if len(tmp) < 3 {
						fmt.Println(nil)
						if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("【使い方】?s show \"ここに任意の個別stampID\" ", false)); err != nil {
							log.Println(err)
							w.WriteHeader(http.StatusInternalServerError)
							return

						}
					}
					var sarchIndex int
					sarchIndex, _ = strconv.Atoi(tmp[2])
					s := strconv.Itoa(array[sarchIndex])
					//fmt.Println(sarchIndex)
					var byte_buf bytes.Buffer
					byte_buf.WriteString("https://stickershop.line-scdn.net/stickershop/v1/sticker/")
					byte_buf.WriteString(s)
					byte_buf.WriteString("/iPhone/sticker@2x.png")

					// 送信元ユーザに注意
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(byte_buf.String(), false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else if strings.Index(event.Text, "?s urlid") == 0 {
					tmp := strings.Split(event.Text, " ")

					if len(tmp) < 3 {
						fmt.Println(nil)
						if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("【使い方】?s urlid \"ここに任意のstampID\" ", false)); err != nil {
							log.Println(err)
							w.WriteHeader(http.StatusInternalServerError)
							return

						}
					}

					s := tmp[2]
					//fmt.Println(sarchIndex)
					var byte_buf bytes.Buffer
					byte_buf.WriteString("http://dl.stickershop.line.naver.jp/products/0/0/1/")
					byte_buf.WriteString(s)
					byte_buf.WriteString("/android/productInfo.meta")

					url := byte_buf.String()

					resp, _ := http.Get(url)
					defer resp.Body.Close()
					//fmt.Println(resp)
					byteArray, _ := ioutil.ReadAll(resp.Body)

					// JSONデコード
					var persons Autogenerated
					if err := json.Unmarshal(byteArray, &persons); err != nil {
						log.Fatal(err)
					}
					// デコードしたデータを表示
					//fmt.Println("取り出した値")

					for _, p := range persons.Stickers {
						fmt.Printf("%d\n", p.ID)
					}
					//fmt.Println("入れた値")
					tmpi := 0
					for i := flag; i < flag+len(persons.Stickers); i++ {
						array[i] = persons.Stickers[tmpi].ID
						fmt.Println(array[i])
						tmpi++
					}

					returnFirstnum := strconv.Itoa(flag)
					//fmt.Println(returnFirstnum)
					returnlastnum := strconv.Itoa(flag + len(persons.Stickers))
					//fmt.Println(returnlastnum)
					flag += len(persons.Stickers)

					var resbuffer bytes.Buffer
					resbuffer.WriteString(returnFirstnum)
					resbuffer.WriteString("~")
					resbuffer.WriteString(returnlastnum)

					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(resbuffer.String(), false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					var test bytes.Buffer
					k := strconv.Itoa(flag - 1)
					test.WriteString("\" ?s show ")
					test.WriteString(k)
					test.WriteString("\"など上記のインデックスで表示できます")
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(test.String(), false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else if strings.Index(event.Text, "?s help") == 0 {
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("", false), message); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else if strings.Index(event.Text, "?s list") == 0 {
					iter := client.Collection("stampID").Documents(ctx)
					if err != nil {
						return
					}
					for {
						doc, err := iter.Next()
						if err == iterator.Done {
							break
						}
						if err != nil {
							return
						}
						fmt.Println(doc.Data())
					}
				}
			}
		}

	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}