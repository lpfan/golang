package main

import (
    "fmt"
    "log"
    "github.com/PuerkitoBio/goquery"
    "gopkg.in/mgo.v2"
)

const domainUrl = "http://bmwclub.ua/"
const targetUrlTemplate = "http://www.bmwclub.ua/forums/60-%D0%9A%D1%83%D0%BF%D0%BB%D1%8F-%D0%9F%D1%80%D0%BE%D0%B4%D0%B0%D0%B6%D0%B0-%D0%91%D0%95%D0%97-%D0%9F%D0%A0%D0%90%D0%92%D0%98%D0%9B"

type  Topic struct {
	Url string
	Title string
	Content string
}

func topicWorker(topicLink string) {
    targetUrl := domainUrl + topicLink
    doc, err := goquery.NewDocument(targetUrl)
    if err != nil {
        log.Fatal(err)
    }

    var topicHeader string
    doc.Find("#postlist h2.title").Each(func(i int, s *goquery.Selection){
        topicHeader = s.Text()
        fmt.Println(topicHeader)
    })

    var topicBody string
    doc.Find("#postlist div.postdetails div.content").Each(func(i int, s *goquery.Selection){
        topicBody = s.Text()
        fmt.Println(topicBody)
    })

    session, mongoErr := mgo.Dial("localhost")
    if mongoErr != nil {
      log.Fatal(err)
    }
    defer session.Close()

    c := session.DB("crawler").C("topics")
    mongoErr = c.Insert(&Topic{targetUrl, topicHeader, topicBody})
    if mongoErr != nil {
      log.Fatal(mongoErr)
    }
    log.Print("Topic processed", targetUrl)
}

func main() {

    for pageNum := 1; pageNum < 2; pageNum++ {
        targetUrl := targetUrlTemplate

        if pageNum >= 2 {
            targetUrl = targetUrlTemplate + "/page" + string(pageNum)
        }

        doc, err := goquery.NewDocument(targetUrl)
        if err != nil {
            log.Fatal(err)
        }

        doc.Find("#threads .threadbit a.title").Each(func(i int, s *goquery.Selection) {
            topicLink, ok := s.Attr("href")
            if ok {
                log.Print("Going to process page", topicLink)
                go topicWorker(topicLink)
            }
        })
    }

}
