package gifhelper

import (
	"image"
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

// GetImageFromDir 讀取此目錄中的所有圖片 (依照檔名來決定順序)
func GetImageFromDir(dirPath string) ([]image.Image, error) {
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

	return ReadImages(imgPaths...), nil
}

func ReadImages(filename ...string) []image.Image {
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

func SaveGIF(filename string, anim *gif.GIF) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	return gif.EncodeAll(file, anim)
}
