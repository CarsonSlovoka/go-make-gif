/*
參考資料:
https://gist.github.com/CarsonSlovoka/dd332e468a045242d8dfd2a4503c810b
https://gist.github.com/nitoyon/10108182cc0c12f54878
*/

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"go-make-gif/meta/gifhelper"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Input struct {
	FrontMatter struct {
		Width      int // 字幕計算位置會用到
		Height     int
		OutputPath string
		Delay      int

		// A LoopCount of 0 means to loop forever.
		// A LoopCount of -1 means to show each frame only once.
		// Otherwise, the animation is looped LoopCount+1 times.
		LoopCount int

		Default Frame
	}
	Datas []*Frame
}

type Frame struct {
	ImgPath string
	IsCC    bool
	Delay   int
	X       int
	Y       int

	// over: 0 會有融合的操作
	// src: 1 原始圖像直接放入
	Op draw.Op
}

func main() {
	var cfgFilepath string
	flag.StringVar(&cfgFilepath, "c", ".gif.manifest", "input the config file path")
	flag.Parse()
	file, err := os.Open(cfgFilepath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = file.Close()
	}()
	input, err := getInput(file)
	if err != nil {
		log.Fatal(err)
	}

	// 依序將每個圖片加入到GIF中
	myGif := &gif.GIF{LoopCount: input.FrontMatter.LoopCount}
	for _, frame := range input.Datas {
		// 取得當前的圖片
		var curImg image.Image
		if d := gifhelper.ReadImages(frame.ImgPath); len(d) == 0 {
			continue
		} else {
			curImg = d[0]
		}

		// 獲得圖片的矩形資料: x,y,w,h
		srcRect := curImg.Bounds()

		// 如果圖片為字幕, 將其初始位置更改至畫面底部且置中的地方
		if frame.IsCC {
			imgWidth := srcRect.Dx()
			imgHeight := srcRect.Dy()
			startX := (input.FrontMatter.Width - imgWidth) / 2
			startY := int(float32(input.FrontMatter.Height-imgHeight) * 0.95)

			// 重新計算srcRect
			srcRect.Min = image.Point{
				X: startX + srcRect.Min.X,
				Y: startY + srcRect.Min.Y,
			}
			srcRect.Max = image.Point{
				X: startX + srcRect.Max.X,
				Y: startY + srcRect.Max.Y,
			}
		}

		dstImg := image.NewRGBA(srcRect) // 創建畫布

		draw.Draw(dstImg, // 輸出的圖片
			srcRect,                         // 指定輸出的x,y,w,h
			curImg, image.Point{X: 0, Y: 0}, // src 從來源圖像的該點位置開始畫起
			frame.Op,
		)

		/*
			// 如果有需要可以取消註解，來查看目前的圖片
			dstFile, err := os.Create("debug.png")
			if err != nil {
				panic(err)
			}
			png.Encode(dstFile, dstImg)
			dstFile.Close()
		*/

		myPalette := image.NewPaletted(srcRect, palette.WebSafe)
		draw.Draw(myPalette, // 輸出的圖片
			myPalette.Rect,      // 輸出到圖片的哪一個位置x,y,w,h
			dstImg, srcRect.Min, // 來源
			frame.Op,
		)

		log.Println("加入圖片: ", frame.ImgPath)
		myGif.Image = append(myGif.Image, myPalette) // // 如果要生成gif檔案，他要吃的是一個image.Platte的格式，如果您之前拿dstImg的內容填入此結構，那麼色調會不對
		myGif.Delay = append(myGif.Delay, frame.Delay)
	}

	// 存檔
	if err = gifhelper.SaveGIF(input.FrontMatter.OutputPath, myGif); err != nil {
		log.Println("無法儲存 GIF:", err)
	} else {
		log.Println("已成功生成 GIF 檔案:", input.FrontMatter.OutputPath)
	}
}

func getInput(buf io.Reader) (*Input, error) {
	cfg := &Input{}

	/*
		buf := bytes.NewBuffer([]byte(`
		---
		{
		    "width": 1920,
		    "height": 1080,
		}
		---
		ImgPath,IsCC,Delay,X,Y,Op
		begin.png,,,,,
				`))
	*/

	// handle frontMatter
	scanner := bufio.NewScanner(buf)
	// 找尋開始標記 ---
	var dataFound bool
	var frontMatter strings.Builder // strings.Builder可以有效地用於連續地組合和構建字串，特別是在需要動態添加字串內容時非常方便

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if dataFound {
				// 遇到第二個 ---，代表結束，停止找尋
				break
			} else {
				// 找到第一個 ---，開始記錄資料
				dataFound = true
			}
		} else if dataFound {
			// 將內容寫入結果變數
			frontMatter.WriteString(line)
			frontMatter.WriteString("\n") // 換行符號
		}
	}
	// result := builder.String()
	if err := json.Unmarshal([]byte(frontMatter.String()), &cfg.FrontMatter); err != nil {
		return cfg, err
	}

	// Handle CSV Data
	scanner.Scan() // 緊接著是CSV的表頭，跳過不處理

	// 處理CSV的內容
	var (
		err error
	)
	for scanner.Scan() {
		row := scanner.Text()
		if row == "" {
			continue
		}
		data := strings.Split(row, ",")
		if len(data) != 6 {
			log.Println("此列無效: ", row)
			continue
		}

		frame := new(Frame)
		if !filepath.IsAbs(data[0]) {
			frame.ImgPath = filepath.Join(cfg.FrontMatter.Default.ImgPath, data[0])
		} else {
			frame.ImgPath = data[0]
		}

		if data[1] == "" {
			frame.IsCC = cfg.FrontMatter.Default.IsCC
		} else {
			frame.IsCC, err = strconv.ParseBool(data[1])
			if err != nil {
				log.Println(row, data[1], err)
				continue
			}
		}

		if data[2] == "" {
			frame.Delay = cfg.FrontMatter.Default.Delay
		} else {
			frame.Delay, err = strconv.Atoi(data[2])
			if err != nil {
				log.Println(row, data[2], err)
				continue
			}
		}

		if data[3] == "" {
			frame.X = cfg.FrontMatter.Default.X
		} else {
			frame.X, err = strconv.Atoi(data[3])
			if err != nil {
				log.Println(row, data[3], err)
				continue
			}
		}

		if data[4] == "" {
			frame.Y = cfg.FrontMatter.Default.Y
		} else {
			frame.Y, err = strconv.Atoi(data[4])
			if err != nil {
				log.Println(row, data[4], err)
				continue
			}
		}

		if data[5] == "" {
			frame.Op = cfg.FrontMatter.Default.Op
		} else {
			var op int
			op, err = strconv.Atoi(data[5])
			if err != nil {
				log.Println(row, data[5], err)
				continue
			}
			frame.Op = draw.Op(op)
		}

		cfg.Datas = append(cfg.Datas, frame)
	}

	// 檢查是否有讀取錯誤
	if err = scanner.Err(); err != nil {
		return cfg, fmt.Errorf("讀取檔案錯誤 %w", err)
	}

	return cfg, nil
}
