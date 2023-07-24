# go-make-gif

一個簡易生成gif的工具

## 🛠️Install

本專案組共提供兩個工具

1. make-gif: 純粹就是把每一張圖片串在一起，只能調整delay的時間而已。
2. make-gifEx: 這個可以提供更多的調整，例如字幕自動放置到底部中間, 可以選擇要截取圖片來源的位置，以及貼到目的地的位置，都能調整

### install make-gif
```yaml
git clone https://github.com/CarsonSlovoka/go-make-gif.git
cd go-make-gif/make-gif
go install # 安裝在 GOPATH/bin 之中
```

###
install make-gif
```yaml
git clone https://github.com/CarsonSlovoka/go-make-gif.git
cd go-make-gif/make-gifex
go install # 安裝在 GOPATH/bin 之中
```

> 請將`GOPATH/bin`將入到系統變數中會比較容易使用
>
> 或者您想要套過`go build -o <xxx>.exe`來產生執行檔也可以

## 🎮USAGE:

前置動作，請先準備好您的圖片集

> 您如果想要快速從剪貼簿蒐集圖片，可以參考我寫的powershell腳本[Watch-ClipboardImage](https://github.com/CarsonSlovoka/powershell/blob/4d30d3137f50e01967ac3d235ded48c8a10a2e0b/src/keyboard/clipboard.psm1#L335-L486)它可以監聽剪貼簿，當剪貼簿有圖片時會自動將圖片保存在指定的目錄

### make-gif.exe

1. 準備好您的圖片集，假設您放在: `C:\...\myImgDir`
2. 運行指令

    ```
    make-gif -imgDir="C:\...\myImgDir" -delay=100 -outputPath="./result.gif"
    ```

### make-gifex.exe

請項目只需要為入一個json檔案，請參考[.gif.manifest.json](src/make-gifex/.gif.manifest.json)

```json5
{
  "width": 173, // 高、寬為cc字幕會用到的內容
  "height": 178,
  "outputPath": "./testFiles/result.gif",
  "loopCount": 2, // 0: loop forever, -1: 每幀圖片只會顯示一次, 其他數值: 重複到該次數為止
  "default": { // 可以統一調整每張圖片的預設值，如果不想額外調整，可以不需要寫
    // 所有可以調整的項目可以參考: https://github.com/CarsonSlovoka/go-make-gif/blob/ed6b8ecf8a06b35af9463c78a1b406c3f36e51ed/src/make-gifex/main.go#L42-L58
    "imgPath": "./testFiles", // 圖片的根目錄，讓您在設定圖片可以不需要寫太多路徑
    "isCC": false, // 若是cc字幕，會自動放到畫面的底部且正中的位置
    "delay": 50 // 1單位為0.01秒，所以50為0.5秒
  },
  "datas": [ // 以下放置您的圖片資訊
    {
      "imgPath": "index.png", // 要使用絕對路徑也可以, 注意第一章圖片最好要大一點，不然會遇到錯誤: gif: image block is out of bounds
      "delay": 0
    },
    {
      "imgPath": "apple.png"
    },
    {
      "imgPath": "banana.png"
    },
    {
      "note:": "",
      "imgPath": "apple_leaf.png",
      "delay": 150,
      "sx": 40, // 來源圖片的x開始位置
      "sy": 40, // 來源圖片的y位置
      "w": 73, // width
      "h": 78, // height
      "dstX": 50, // 目的地的x位置
      "dstY": 20 // destinationY
      // 沒有dstWidth: 寬高都用和來源圖片一模一樣的大小
    },
    {
      "imgPath": "myCC.png",
      "isCC": true // 此時的位置會自動調整，不需要設定位置
    },
    {
      "imgPath": "end.png"
    }
  ]
}
```

> 有關產生cc字幕的圖片，可以考慮使用簡單的html來產生, 可以參考[make-cc.html](tool/make-cc.html)，直接打開敲入想要的字幕就能用生成

## 疑難排解

### gif: image block is out of bounds

請確保之前的圖片尺寸至少有一張比目前的尺寸要大
