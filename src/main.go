package main

import (
	"flag"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png" // 如果你的格式是這些，就要import相關包，不然image.Decode會錯誤
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	imgSlice, err := getImageFromDir(cfg.imgDir)
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
	if err := saveGIF(cfg.outputPath, myGif); err != nil {
		log.Println("無法儲存 GIF:", err)
		return
	}
	log.Println("已成功生成 GIF 檔案:", cfg.outputPath)
}

// 讀取此目錄中的所有圖片 (依照檔名來決定順序)
func getImageFromDir(dirPath string) ([]image.Image, error) {
	var imgPaths []string
	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		pathLower := strings.ToLower(path)
		if info.IsDir() || (!strings.HasSuffix(pathLower, "jpeg") && !strings.HasSuffix(pathLower, "png")) {
			return nil
		}
		imgPaths = append(imgPaths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(imgPaths) // 依照名稱來排序, []string{"44", "33", "2", "32", "1"} => [1 2 32 33 44]

	return readImages(imgPaths...), nil
}

func readImages(filename ...string) []image.Image {
	var imgSlice []image.Image
	for _, filePath := range filename {
		file, err := os.Open(filePath)
		if err != nil {
			log.Println("[error]", err)
			continue
		}

		img, _, err := image.Decode(file)
		_ = file.Close()
		if err != nil {
			log.Println("[error]", err)
			continue
		}
		imgSlice = append(imgSlice, img)
	}
	return imgSlice
}

func saveGIF(filename string, anim *gif.GIF) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	return gif.EncodeAll(file, anim)
}
