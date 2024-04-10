package main

import (
	"fmt"
	"os"

	"image/jpeg"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/ini.v1"

	"github.com/kbinani/screenshot"
)

func main() {

	// アプリエラー発生時に、画面が閉じてしまうと原因がわからないので
	// 画面に表示させるために利用したり
	// Discordからの通信待ちをするのに利用します。
	stopProgram := make(chan bool)

	// iniファイルからDiscordeのBOT設定値を読み込む
	discordSetting, err := loadSetting()
	if err != nil {
		fmt.Println(err)
		fmt.Println("設定ファイルからTokenが取得できませんでした。")
		fmt.Println("設定ファイルをご確認の上、この画面を×で閉じてから再度実行してください。")
		<-stopProgram
	}

	//Discordのセッションを作成
	discord, err := discordgo.New()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Discordのセッションを作成できませんでした。")
		fmt.Println("ここでは特に何もしていないはずなので、エラーの原因がわかりませんでした。")
		fmt.Println("設定ファイル等をご確認の上、この画面を×で閉じてから再度実行してください。")
		<-stopProgram
	}

	// トークンを設定
	discord.Token = discordSetting.Token()

	// Discordeから受信したイベントを受け取る
	// ここでいうイベントとは会話
	discord.AddHandler(onMessageCreate)

	// discordeのBotとの通信を開始
	err = discord.Open()
	if err != nil {
		fmt.Println(err)
		fmt.Println("DiscordのBotとの通信を開始できませんでした。")
		fmt.Println("Tokenの設定、Bot側の設定を間違えている可能性があります")
		fmt.Println("設定状況をご確認の上、この画面を×で閉じてから再度実行してください。")
		<-stopProgram
	}

	// アプリ終了時にdiscordeのBotとの通信を終了させる処理
	defer discord.Close()

	fmt.Println("プログラムスタート！")

	//プログラムが終了しないようロック
	// stopBotに通信が来るまでここで待機という書き方。
	// 送信ロジックを作っていないので、一生勝手には進まない。

	<-stopProgram

}

// Discordから受信したメッセージを処理します。
func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// しゃべったのがbotの場合には処理をしません。
	if m.Author.Bot {
		return
	}

	// スクリーンショットを撮って画像ファイルに保存
	err := createNewScreenshotJpeg()
	if err != nil {
		fmt.Println("スクリーンショットの画像保存に失敗しました。")
		return
	}
	//Discord側に画像を送信
	sendfile(s, m.ChannelID, "out.png") //画像送信対応

}

// ///////////////////
// 画像送信
// ///////////////////
func sendfile(s *discordgo.Session, channelID string, jpegName string) error {

	screenshotJpeg, err := os.Open("out.png")
	if err != nil {
		return fmt.Errorf("送信する画像が見当たりません。err:%v", err.Error())
	}
	defer screenshotJpeg.Close()
	_, err = s.ChannelFileSend(channelID, jpegName, screenshotJpeg)

	if err != nil {
		return fmt.Errorf("画像をDiscordeに送信したけど、エラーが返却されたよ err:%v", err.Error())
	}
	return nil
}

// ///////////////////
// スクショand保存
// ///////////////////
func createNewScreenshotJpeg() error {
	bounds := screenshot.GetDisplayBounds(0)
	img, _ := screenshot.CaptureRect(bounds)
	file, err := os.Create("out.png")
	if err != nil {
		return fmt.Errorf("スクリーンショットの画像保存に失敗しました。 err:%v", err.Error())
	}
	defer file.Close()
	jpeg.Encode(file, img, &jpeg.Options{Quality: 80})

	return nil
}

// ///////////////////
// ini からToken情報を取得してくる処理
// ///////////////////
type DiscordSetting struct {
	tokenId string
}

func newDiscordSetting(token string) DiscordSetting {
	return DiscordSetting{token}
}

func (d DiscordSetting) Token() string {
	return d.tokenId
}

func loadSetting() (DiscordSetting, error) {
	// ファイル読み込み
	cfg, err := ini.Load("discordSetting.ini")
	if err != nil {

		return DiscordSetting{}, fmt.Errorf("設定ファイル(discordSetting.ini)が、見当たりません。 err:%v", err)
	}

	// iniからtokenidを取得
	tokenId := cfg.Section("bot").Key("token").String()

	_tokenId := "Bot " + tokenId

	return newDiscordSetting(_tokenId), nil
}

////////////////////////
