package main

import (
    "fmt"
    "log"
    "github.com/PuerkitoBio/goquery"
    "gopkg.in/mgo.v2"
    "golang.org/x/text/transform"
    "golang.org/x/text/encoding/charmap"
    "strings"
    "io/ioutil"
    "time"
)

const domainUrl = "http://bmwclub.ua/"
const targetUrlTemplate = "http://www.bmwclub.ua/forums/60-%D0%9A%D1%83%D0%BF%D0%BB%D1%8F-%D0%9F%D1%80%D0%BE%D0%B4%D0%B0%D0%B6%D0%B0-%D0%91%D0%95%D0%97-%D0%9F%D0%A0%D0%90%D0%92%D0%98%D0%9B"

type  Topic struct {
	Url string
	Title string
	Content string
  CreatedAt time.Time
}

func decodeStringToUtf(rawString string) string {
  sr := strings.NewReader(rawString)
  tr := transform.NewReader(sr, charmap.Windows1251.NewDecoder())
  buf, err := ioutil.ReadAll(tr)
  if err != nil {
    log.Fatal(err)
  }
  return string(buf)
}

func topicWorker(s *mgo.Session, topicChannel<-chan string) {
  for {
    select {
      case topicLink := <- topicChannel:
        log.Printf("Recieved %s for processing", topicLink)
        targetUrl := domainUrl + topicLink
        doc, err := goquery.NewDocument(targetUrl)
        if err != nil {
          log.Fatal(err)
        }
        targetUrl = decodeStringToUtf(targetUrl)

        var topicHeader string
        topicHeader = doc.Find("div#postlist #posts div.postdetails div.postbody div.postrow h2.title").First().Text()
        topicHeader = strings.TrimSpace(topicHeader)
        topicHeader = decodeStringToUtf(topicHeader)
        fmt.Println(topicHeader)

        var topicBody string
        topicBody = doc.Find("div#postlist #posts div.postdetails div.postbody div.postrow div.content").First().Text()
        topicBody = decodeStringToUtf(strings.TrimSpace(topicBody))

        var topicCreatedAt string
        topicCreatedAt = doc.Find("div#postlist #posts .postcontainer div.posthead .postdate").First().Text()
        topicCreatedAt = strings.TrimSpace(topicCreatedAt)
        topicCreatedAt = decodeStringToUtf(topicCreatedAt)

        const layout = "02.01.2006,В 15:04"
        topicCreatedAtTime, timeErr := time.Parse(layout, topicCreatedAt)
        if timeErr != nil {
          log.Print(timeErr)
        }

        session := s.Copy()
        defer session.Close()
        c := session.DB("crawler").C("topics")

        mongoErr := c.Insert(&Topic{targetUrl, topicHeader, topicBody, topicCreatedAtTime})
        if mongoErr != nil {
          log.Fatal(mongoErr)
        }
        log.Print("Processing next topic", targetUrl)
      default:
        log.Print("No links")
      }
  }
}

func main() {
    var topicChannel = make(chan string)


    var topics []string
    for pageNum := 1; pageNum <= 5; pageNum++ {
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
                topics = append(topics, topicLink)
            }
        })
    }
    session, mongoErr := mgo.Dial("localhost")
    if mongoErr != nil {
      log.Fatal(mongoErr)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)

    for wCount := 0; wCount < 30; wCount++ {
      go topicWorker(session, topicChannel)
    }

    for _, topic := range topics {
      topicChannel <- topic
    }

}
