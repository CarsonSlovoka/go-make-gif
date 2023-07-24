/*
參考資料:
https://gist.github.com/CarsonSlovoka/dd332e468a045242d8dfd2a4503c810b
https://gist.github.com/nitoyon/10108182cc0c12f54878
*/

package main

import (
	"encoding/json"
	"flag"
	. "github.com/CarsonSlovoka/go-pkg/v2/op"
	"go-make-gif/meta/gifhelper"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"os"
	"path/filepath"
)

type Input struct {
	FrontMatter
	Datas []*Frame
}

type FrontMatter struct {
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

type Frame struct {
	// Phone int    `json:"phone,omitempty"`
	ImgPath string
	IsCC    bool
	Delay   *int // 1unit=0.01sec

	Sx   *int
	Sy   *int
	W    *int
	H    *int
	DstX *int
	DstY *int

	// over: 0 會有融合的操作
	// src: 1 原始圖像直接放入
	Op *draw.Op
}

func (f *Frame) setDefaults(d *Frame, rect image.Rectangle) {
	if f.Delay == nil {
		dDelay := 100
		f.Delay = If(d.Delay != nil, d.Delay, &dDelay)
	}
	if f.Sx == nil {
		f.Sx = If(d.Sx != nil, d.Sx, &rect.Min.X)
	}
	if f.Sy == nil {
		f.Sy = If(d.Sy != nil, d.Sy, &rect.Min.Y)
	}
	if f.Op == nil {
		dOp := draw.Src
		f.Op = If(d.Op != nil, d.Op, &dOp)
	}
}

func main() {
	var cfgFilepath string
	flag.StringVar(&cfgFilepath, "c", ".gif.manifest.json", "input the config file path")
	flag.Parse()
	input := Input{}
	if file, err := os.Open(cfgFilepath); err != nil {
		log.Fatal(err)
	} else {
		if err = json.NewDecoder(file).Decode(&input); err != nil {
			_ = file.Close()
			log.Fatal(err)
		}
		_ = file.Close()
	}

	// 依序將每個圖片加入到GIF中
	myGif := &gif.GIF{LoopCount: input.FrontMatter.LoopCount}
	for _, frame := range input.Datas {
		if !filepath.IsAbs(frame.ImgPath) {
			frame.ImgPath = filepath.Join(input.FrontMatter.Default.ImgPath, frame.ImgPath)
		}
		// 取得當前的圖片
		curImg, err := readImage(frame.ImgPath)
		if err != nil {
			log.Println(err)
			continue
		}

		// 獲得圖片的矩形資料: x,y,w,h
		dstRect := curImg.Bounds()

		frame.setDefaults(&input.FrontMatter.Default, dstRect)

		// 如果圖片為字幕, 將其初始位置更改至畫面底部且置中的地方
		if frame.IsCC {
			// 如果是字幕, sx, sy, w, h就依據實際的照片尺寸，不做任何調整
			sX := 0
			sY := 0
			frame.Sx = &sX
			frame.Sy = &sY

			w := dstRect.Dx()
			y := dstRect.Dy()
			frame.W = &w
			frame.H = &y

			dstX := (input.FrontMatter.Width - *frame.W) / 2
			frame.DstX = &dstX
			dstY := int(float32(input.FrontMatter.Height-*frame.H) * 0.95)
			frame.DstY = &dstY

			// 重新計算dstRect
			dstRect.Min = image.Point{
				X: *frame.DstX + dstRect.Min.X,
				Y: *frame.DstY + dstRect.Min.Y,
			}
			dstRect.Max = image.Point{
				X: *frame.DstX + dstRect.Max.X,
				Y: *frame.DstY + dstRect.Max.Y,
			}
		} else { // 非cc字幕
			if frame.W == nil {
				width := dstRect.Dx() - *frame.Sx // 扣除sX才不會越界，不然越界後的內容都是黑的
				frame.W = &width
			} else { // 自動幫忙計算最大寬度，避免寬度超過，導致的黑色內容產生
				if maxW := dstRect.Dx() - *frame.Sx; *frame.W > maxW {
					frame.W = &maxW
				}
			}
			if frame.H == nil {
				height := dstRect.Dy() - *frame.Sy
				frame.H = &height
			} else {
				if maxH := dstRect.Dy() - *frame.Sy; *frame.H > maxH {
					frame.H = &maxH
				}
			}

			if frame.DstX == nil {
				dstX := dstRect.Min.X // 用原圖片的sx當作起始位置
				frame.DstX = &dstX
			}
			if frame.DstY == nil {
				dstY := dstRect.Min.Y
				frame.DstY = &dstY
			}

			// 重新計算dstRect
			dstRect.Min = image.Point{
				X: *frame.DstX,
				Y: *frame.DstY,
			}
			dstRect.Max = image.Point{
				X: *frame.DstX + *frame.W,
				Y: *frame.DstY + *frame.H,
			}
		}

		log.Println("[debug]", filepath.Base(frame.ImgPath), dstRect.Min.X, dstRect.Min.Y, dstRect.Dx(), dstRect.Dy())
		dstImg := image.NewRGBA(dstRect) // 創建畫布

		draw.Draw(dstImg, // 輸出的圖片
			dstRect,                                         // 指定輸出的x,y,w,h
			curImg, image.Point{X: *frame.Sx, Y: *frame.Sy}, // src 從來源圖像的該點位置開始畫起
			*frame.Op,
		)

		//
		//	// 如果有需要可以取消註解，來查看目前的圖片
		//	dstFile, err := os.Create("debug.png")
		//	if err != nil {
		//		panic(err)
		//	}
		//	png.Encode(dstFile, dstImg)
		//	dstFile.Close()
		//

		myPalette := image.NewPaletted(dstRect, palette.WebSafe)
		draw.Draw(myPalette, // 輸出的圖片
			myPalette.Rect,      // 輸出到圖片的哪一個位置x,y,w,h
			dstImg, dstRect.Min, // 來源
			*frame.Op,
		)

		log.Println("加入圖片: ", frame.ImgPath)
		myGif.Image = append(myGif.Image, myPalette) // // 如果要生成gif檔案，他要吃的是一個image.Platte的格式，如果您之前拿dstImg的內容填入此結構，那麼色調會不對
		myGif.Delay = append(myGif.Delay, *frame.Delay)
	}

	// 存檔
	if err := gifhelper.SaveGIF(input.FrontMatter.OutputPath, myGif); err != nil {
		log.Println("無法儲存 GIF:", err)
	} else {
		log.Println("已成功生成 GIF 檔案:", input.FrontMatter.OutputPath)
	}

}

func readImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, err
}
