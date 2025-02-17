package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"golang.org/x/sync/errgroup"
)

func downloadRange(ctx context.Context, url string, start, end int64, wg *sync.WaitGroup, outFile *os.File, g *errgroup.Group) {
	defer wg.Done()

	// HTTPリクエスト作成
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		g.Go(func() error { return err })
		return
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	// HTTPリクエスト送信
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		g.Go(func() error {return err})
		return
	}
	defer resp.Body.Close()

	// 書き込み
	_, err = outFile.Seek(start, io.SeekStart)
	if err != nil {
		g.Go(func() error { return err })
		return
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		g.Go(func() error { return err })
		return
	}
}

func main() {
	// ダウンロードするファイルのURL
	url := "http://example.com/largefile"

	// ダウンロードするファイルのサイズを取得(仮定)
	fileSize := int64(1000000) // 1MBのファイル

	// ゴルーチンのエラー用
	var g errgroup.Group

	// 出力ファイルの作成
	outFile, err := os.Create("downloaded_file")
	if err != nil {
		fmt.Println("出力ファイルエラー:", err)
		return
	}
	defer outFile.Close()

	// ダウンロード範囲の分割
	numParts := 4
	partSize := fileSize / int64(numParts)

	// ゴルーチンを使って並行ダウンロードする
	var wg sync.WaitGroup
	for i := 0; i < numParts; i++ {
		start := int64(i) * partSize
		end := start + partSize - 1
		if i == numParts - 1 {
			end = fileSize - 1
		}

		wg.Add(1)
		go downloadRange(context.Background(), url, start, end, &wg, outFile, &g)
	} 
	// エラーが発生した場合に処理をキャンセル
	if err := g.Wait(); err != nil {
		fmt.Println("ダウンロード中にエラー:", err)
		return
	}
	
	wg.Wait()
	fmt.Println("ダウンロード完了！")
}

