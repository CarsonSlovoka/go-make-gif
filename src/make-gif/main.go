package main

import (
	"flag"
	"go-make-gif/meta/gifhelper"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
)

type config struct {
	imgDir     string
	delay      int
	outputPath string
}

func parseConfig() *config {
	cfg := &config{}
	flag.StringVar(&cfg.imgDir, "imgDir", ".", "the image source of the directory")
	flag.IntVar(&cfg.delay, "delay", 100, "millisecond. default: 0.1sec")
	flag.StringVar(&cfg.outputPath, "outputPath", "result.gif", "output filepath")
	flag.Parse()
	return cfg
}

func main() {
	// 讀取設定參數
	cfg := parseConfig()

	// 讀取圖片集
	imgSlice, err := gifhelper.GetImageFromDir(cfg.imgDir)
	if err != nil {
		log.Fatal(err)
	}

	// 依序將每個圖片加入到GIF中
	myGif := &gif.GIF{}
	for _, curImg := range imgSlice {
		bound := curImg.Bounds()
		myPalette := image.NewPaletted(bound, palette.WebSafe)            // 調色盤選擇, WebSafe(生成的文件會比Plan9小一點,顏色個人也覺得這比較好看), Plan9
		draw.Draw(myPalette, myPalette.Rect, curImg, bound.Min, draw.Src) // 要畫才會真的有內容

		myGif.Image = append(myGif.Image, myPalette)
		myGif.Delay = append(myGif.Delay, 100) // 每一幀的時間間隔，單位為10ms (這裡設置為100ms)
	}

	// 存檔
	if err = gifhelper.SaveGIF(cfg.outputPath, myGif); err != nil {
		log.Println("無法儲存 GIF:", err)
		return
	}
	log.Println("已成功生成 GIF 檔案:", cfg.outputPath)
}
