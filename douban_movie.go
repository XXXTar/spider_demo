package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	USERNAME = "root"
	PASSWORD = "Xxy20030607"
	HOST     = "localhost"
	PORT     = "3306"
	DBNAME   = "douban_movie"
)

var DB *sql.DB

type MovieData struct {
	Title    string `json:"title"`
	Director string `json:"director"`
	Picture  string `json:"picture"`
	Actor    string `json:"actor"`
	Year     string `json:"year"`
	Score    string `json:"score"`
	Quote    string `json:"quote"`
}

func main() {
	InitDB()
	for i := 0; i < 10; i++ {
		// fmt.Printf("正在爬取第%d页信息", i)
		Spider(strconv.Itoa(i * 25))
	}

}

func Spider(page string) {

	//发送请求
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://movie.douban.com/top250?start="+page, nil) //get一般没有请求体
	if err != nil {
		fmt.Println("req err", err)
	}
	//防止浏览器检测爬虫访问，所以加一些请求头
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 SLBrowser/9.0.3.1311 SLBChan/103")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Referer", "https://movie.douban.com/chart")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求失败", err)
	}
	defer resp.Body.Close()
	//解析网页
	docDetail, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("解释失败", err)
	}
	//获取节点
	//#content > div > div.article > ol > li:nth-child(1) > div > div.info > div.hd > a > span:nth-child(1)
	//#content > div > div.article > ol > li:nth-child(1) > div > div.pic > a > img
	docDetail.Find("#content > div > div.article > ol > li"). //返回一个列表
									Each(func(i int, selection *goquery.Selection) { //在列表中继续找
			var data MovieData
			title := selection.Find("div > div.info > div.hd > a > span:nth-child(1)").Text()
			img := selection.Find("div > div.pic > a > img") //img 标签-->src属性里面
			imgTmp, ok := img.Attr("src")
			info := selection.Find("div > div.info > div.bd > p:nth-child(1)").Text()
			score := selection.Find("div > div.info > div.bd > div > span.rating_num").Text()
			quote := selection.Find("div > div.info > div.bd > p.quote > span").Text()
			if ok {
				director, actor, year := InfoSpite(info)
				data.Title = title
				data.Director = director
				data.Picture = imgTmp
				data.Actor = actor
				data.Year = year
				data.Score = score
				data.Quote = quote
				//保存信息
				if InsertData(data) {

				} else {
					fmt.Println("插入失败")
					return
				}

				// fmt.Println("data", data)
			}
		})
	fmt.Println("插入成功")
}

func InfoSpite(info string) (director, actor, year string) {
	directorRe, _ := regexp.Compile(`导演:(.*)主演:`)
	director = string(directorRe.Find([]byte(info)))
	actorRe, _ := regexp.Compile(`主演:(.*)`)
	actor = string(actorRe.Find([]byte(info)))
	yearRe, _ := regexp.Compile(`(\d+)`)
	year = string(yearRe.Find([]byte(info)))
	return
}

func InitDB() {
	path := strings.Join([]string{USERNAME, ":", PASSWORD, "@tcp(", HOST, ":", PORT, ")/", DBNAME, "?charset=utf8"}, "")
	DB, _ = sql.Open("mysql", path)
	DB.SetConnMaxLifetime(10)
	DB.SetMaxIdleConns(5)
	if err := DB.Ping(); err != nil {
		fmt.Println("open database fail")
		return
	}
	fmt.Println("connect success")
}

func InsertData(m MovieData) bool {
	tx, err := DB.Begin()
	if err != nil {
		fmt.Println("begin err", err)
		return false
	}
	stmt, err := tx.Prepare("INSERT INTO movie_data (`Title`,`Director`,`Picture`,`Actor`,`Year`,`Score`,`Quote`)" +
		"VALUES(?,?,?,?,?,?,?)")
	if err != nil {
		fmt.Println("Prepare err", err)
		return false
	}
	_, err = stmt.Exec(m.Title, m.Director, m.Picture, m.Actor, m.Year, m.Score, m.Quote)
	if err != nil {
		fmt.Println("exec fail", err)
		return false
	}
	tx.Commit()
	return true
}
