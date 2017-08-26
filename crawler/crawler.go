package main

import (
    "log"
    "github.com/PuerkitoBio/goquery"
    "gopkg.in/mgo.v2"
    "golang.org/x/text/transform"
    "golang.org/x/text/encoding/charmap"
    "strings"
    "io/ioutil"
    "time"
    "sync"
    "strconv"
)

const domainUrl = "http://www.bmwclub.ua/"
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

func topicWorker(s *mgo.Session, wg *sync.WaitGroup, topicLink string) {
        defer wg.Done()

        targetUrl := domainUrl + topicLink
        doc, err := goquery.NewDocument(targetUrl)
        if err != nil {
          log.Fatal(err)
        }

        var topicHeader string
        topicHeader = doc.Find("div#postlist #posts div.postdetails div.postbody div.postrow h2.title").First().Text()
        topicHeader = strings.TrimSpace(topicHeader)
        topicHeader = decodeStringToUtf(topicHeader)

        var topicBody string
        topicBody = doc.Find("div#postlist #posts div.postdetails div.postbody div.postrow div.content").First().Text()
        topicBody = decodeStringToUtf(strings.TrimSpace(topicBody))

        var topicCreatedAt string
        topicCreatedAt = doc.Find("div#postlist #posts .postcontainer div.posthead .postdate").First().Text()
        topicCreatedAt = strings.TrimSpace(topicCreatedAt)
        topicCreatedAt = decodeStringToUtf(topicCreatedAt)

        const layout = "02.01.2006,В 15:04"
        var topicCreatedAtTime time.Time
        topicCreatedAtTime, timeErr := time.Parse(layout, topicCreatedAt)
        if timeErr != nil {
          topicCreatedAtTime = time.Now().Local()
          log.Print(timeErr)
        }

        session := s.Copy()
        defer session.Close()
        c := session.DB("crawler").C("topics")

        mongoErr := c.Insert(&Topic{targetUrl, topicHeader, topicBody, topicCreatedAtTime})
        if mongoErr != nil {
          log.Fatal(mongoErr)
        }
}

func main() {
    var wg sync.WaitGroup

    session, mongoErr := mgo.Dial("localhost")
    if mongoErr != nil {
      log.Fatal(mongoErr)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)

    for pageNum := 1; pageNum <= 5; pageNum++ {
        targetUrl := targetUrlTemplate

        if pageNum >= 2 {
            targetUrl = targetUrlTemplate + "/page" + strconv.Itoa(pageNum)
        }

        doc, err := goquery.NewDocument(targetUrl)
        if err != nil {
            log.Fatal(err)
        }

        doc.Find("div#threadlist div ol#threads .threadbit div.threadinfo div.inner h3.threadtitle a.title").Each(func(i int, s *goquery.Selection) {
          topicLink, ok := s.Attr("href")
          if ok {
            topicLink = decodeStringToUtf(topicLink)
            wg.Add(1)
            go topicWorker(session, &wg, topicLink)
          }
        })
        //topicLink, ok := doc.Find("div#threadlist div ol#threads .threadbit div.threadinfo div.inner h3.threadtitle a.title").First().Attr("href")
        //if ok {
        //  topicLink = decodeStringToUtf(topicLink)
        //  wg.Add(1)
        //  go topicWorker(session, &wg, topicLink)
        //}
        log.Printf("Processed %d page", pageNum)
    }
    wg.Wait()
}
