package main

import (
    "fmt"
    "log"
    "github.com/PuerkitoBio/goquery"
)

const targetUrl = "http://www.bmwclub.ua/forums/60-%D0%9A%D1%83%D0%BF%D0%BB%D1%8F-%D0%9F%D1%80%D0%BE%D0%B4%D0%B0%D0%B6%D0%B0-%D0%91%D0%95%D0%97-%D0%9F%D0%A0%D0%90%D0%92%D0%98%D0%9B"

func main() {
    doc, err := goquery.NewDocument(targetUrl)
    if err != nil {
        log.Fatal(err)
    }

    doc.Find("#threads .threadbit a.title").Each(func(i int, s *goquery.Selection) {
        topicLink, ok := s.Attr("href")
        if ok {
            fmt.Println(topicLink)
        }
    })
}
