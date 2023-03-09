package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
)

// https://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/weather.html#response_field
type WeatherResponse struct {
	ResultInfo struct {
		Count       int     `json:"Count"`
		Total       int     `json:"Total"`
		Start       int     `json:"Start"`
		Status      int     `json:"Status"`
		Latency     float64 `json:"Latency"`
		Description string  `json:"Description"`
		Copyright   string  `json:"Copyright"`
	} `json:"ResultInfo"`
	Feature []struct {
		ID       string `json:"Id"`
		Name     string `json:"Name"`
		Geometry struct {
			Type        string `json:"Type"`
			Coordinates string `json:"Coordinates"`
		} `json:"Geometry"`
		Property struct {
			WeatherAreaCode int `json:"WeatherAreaCode"`
			WeatherList     struct {
				Weather []struct {
					Type     string      `json:"Type"`
					Date     string      `json:"Date"`
					Rainfall json.Number `json:"Rainfall"`
				} `json:"Weather"`
			} `json:"WeatherList"`
		} `json:"Property"`
	} `json:"Feature"`
}

var (
	SLACK_TOKEN, CHANNEL_ID, YAHOO_CLIENT_ID, LNG_LAT string
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Failed to load env:", err)
	}
	SLACK_TOKEN = os.Getenv("BOT_USER_OAUTH_TOKEN")
	CHANNEL_ID = os.Getenv("CHANNEL_ID")
	YAHOO_CLIENT_ID = os.Getenv("YAHOO_CLIENT_ID")
	LNG_LAT = os.Getenv("LNG_LAT")
}

func main() {
	c := cron.New()
	c.AddFunc("@every 10m", sendToSlack)
	c.Start()

	select {}
}

func sendToSlack() {
	// YahooAPIから天気取得
	wr, err := fetchWeather()
	if err != nil {
		fmt.Println("Failed to fetch weather info from yahoo api:", err)
	}
	if !isSendable(wr) {
		text := createWeatherTable(wr)
		now, _ := wr.Feature[0].Property.WeatherList.Weather[0].Rainfall.Float64()
		if now == 0 {
			text += "<!channel> 60分間の間に雨が降る恐れがあります"
		} else {
			text += "<!channel> 雨が止みます"
		}
		c := slack.New(SLACK_TOKEN)
		// MsgOptionText() 第二引数: 特殊文字をエスケープするかどうか
		_, _, err = c.PostMessage(CHANNEL_ID, slack.MsgOptionText(text, false))
		if err != nil {
			fmt.Println("Failed to post message:", err)
		}
	}
}

func fetchWeather() (wr *WeatherResponse, err error) {
	url := "https://map.yahooapis.jp/weather/V1/place?output=json&" + "coordinates=" + LNG_LAT + "&appid=" + YAHOO_CLIENT_ID
	res, _ := http.Get(url)
	byteArray, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	jsonBytes := ([]byte)(byteArray)
	fmt.Println(string(jsonBytes))
	wh := new(WeatherResponse)
	if err := json.Unmarshal(jsonBytes, wh); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return nil, err
	}
	return wh, nil
}

/*
投稿フォーマット

	*地点(xxx.xxxxxx,xx.xxxxxx)の2023年03月03日 15時50分から60分間の天気情報*

	```
	時間      : 降水強度(mm/h)
	xx時xx分  : x.x
	xx時xx分  : x.x
	xx時xx分  : x.x
	xx時xx分  : x.x
	xx時xx分  : x.x
	xx時xx分  : x.x
	xx時xx分  : x.x
	```
*/
func createWeatherTable(wr *WeatherResponse) string {
	var text string
	text += "*" + wr.Feature[0].Name + "*" + "\n"
	text += "```\n時間      : 降水強度(mm/h)" + "\n"
	for _, v := range wr.Feature[0].Property.WeatherList.Weather {
		text += formatDate(v.Date) + "  : " + v.Rainfall.String() + "\n"
	}
	text += "```\n"
	return text
}

func isSendable(wr *WeatherResponse) bool {
	now, _ := wr.Feature[0].Property.WeatherList.Weather[0].Rainfall.Float64()
	for _, v := range wr.Feature[0].Property.WeatherList.Weather[1:] {
		if n, _ := v.Rainfall.Float64(); n > 0.0 && now == 0.0 {
			// 雨が降るパターン
			return true
		} else if n > 0.0 && now > 0.0 {
			return false
		}
	}
	// 雨が止むかどうか
	return now > 0.0
}

func formatDate(d string) string {
	layout := "200601021504"
	t, err := time.Parse(layout, d)
	if err != nil {
		fmt.Println("Format error:", err)
		return ""
	}
	return t.Format("15時04分")
}
