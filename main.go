package main

import (
	"bufio"
	"log"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/yanyiwu/gojieba"
)

type Sentence struct {
	Sent    string `json:"Sent"`
	SegWord []string `json:"Segword"`
}


var jieba *gojieba.Jieba

func main() {
	jieba = gojieba.NewJieba()
	defer jieba.Free()
	r:=setupRouter()
	r.Run()

}

func setupRouter() *gin.Engine{
	r := gin.Default()
	r.POST("/add_newword",func(ctx *gin.Context) {
		word:=ctx.PostForm("Word")
		jieba.AddWord(word)
		ctx.JSON(http.StatusOK,"succ add new word")
	})
	r.POST("/get_segword", func(c *gin.Context) {
		file,err:=c.FormFile("file")
		if err!=nil{
			c.JSON(http.StatusOK,gin.H{
				"msg":"get file fail",
				"err":err,
			})
			return
		}
		f,err:=file.Open()
		if err!=nil{
			c.JSON(http.StatusOK,gin.H{
				"msg":"open file fail",
			})
			return
		}
		defer f.Close()
		result := make(chan Sentence, 1000)
		go scanFile(f, result)
		ret := make([]Sentence, 0, 200000)
		for v := range result {
			ret = append(ret, v)
		}
		c.JSON(http.StatusOK,ret)
	})
	return r
}

func scanFile(file multipart.File, result chan<- Sentence) {

	jobs := make(chan string, 1000)
	//开启协程
	wg := new(sync.WaitGroup)
	for w := 1; w <= 5; w++ {
		wg.Add(1)
		go GetSegWord(jobs, result, wg)
	}
	//按行读取
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		if sc.Text() != "" {
			jobs<-sc.Text()
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan file error: %v", err)
		return
	}
	close(jobs)
	wg.Wait()
	close(result)
}

func GetSegWord(job <-chan string, result chan<- Sentence, wg *sync.WaitGroup) {
	defer wg.Done()

	
	for {
		v, ok := <-job
		if !ok {
			break
		}
		S := Sentence{Sent: v, SegWord:jieba.CutAll(v)}

		result <- S
	}
}
