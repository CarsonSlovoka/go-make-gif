# go-make-gif

一個簡易生成gif的工具

## Install

```yaml
git clone https://github.com/CarsonSlovoka/go-make-gif.git
cd go-make-gif/make-gif
go install # 安裝在 GOPATH/bin 之中
```

> 請將`GOPATH/bin`將入到系統變數中會比較容易使用

## USAGE

1. 準備好您的圖片集(假設您放在: `C:\...\myImgDir`)
    > 您如果想要快速從剪貼簿蒐集圖片，可以參考我寫的powershell腳本[Watch-ClipboardImage](https://github.com/CarsonSlovoka/powershell/blob/4d30d3137f50e01967ac3d235ded48c8a10a2e0b/src/keyboard/clipboard.psm1#L335-L486)它可以監聽剪貼簿，當剪貼簿有圖片時會自動將圖片保存在指定的目錄

3. 運行指令

    ```
    make-gif -imgDir="C:\...\myImgDir" -delay=100 -outputPath="./result.gif"
    ```
