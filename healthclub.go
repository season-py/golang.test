package heathclub

import (
	"io"
	"fmt"
	"net"
	"time"
	"sync"
	"strings"
	"net/url"
	"math/rand"
	"regexp"
	"strconv"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

var domain = "http://10.5.17.74"

var users = [] User {{"haishanzhang@cyou-inc.com", 
		 			  "CC6E08BB1B6FEF7BA0BAB980407D4AA74C36DB72FA6DCD7A796DFED5F4DCA9C7DCD4455981C442168A4B4784A4F236E89748D47B3E5504DFD9C1EAB9960D5B55D8C369093FF16619", 
		 			  "ccxmhOEyPJnULPkMmzmTT+QpnmI=",
		 			  [] int{0, 1, 0, 1, 0, 0, 0}}, 
		 			 {"haishanzhang@cyou-inc.com", 
		 			  "CC6E08BB1B6FEF7BA0BAB980407D4AA74C36DB72FA6DCD7A796DFED5F4DCA9C7DCD4455981C442168A4B4784A4F236E89748D47B3E5504DFD9C1EAB9960D5B55D8C369093FF16619", 
		 			  "ccxmhOEyPJnULPkMmzmTT+QpnmI=",
		 			  [] int{0, 0, 0, 0, 0, 0, 0}},
		 			}

var activities = make(chan Activity)

var wg sync.WaitGroup

type User struct {
	email string
	spbforms string
	sbvcode string
	wanna [] int
}

type Activity struct {
	subject string
	date time.Time
	weekday time.Weekday
	ppcnt int
	form_addr string
}

func randint(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func dialTimeout(network, addr string) (net.Conn, error) {
	timeout := time.Duration(5 * time.Second)
	return net.DialTimeout(network, addr, timeout)
}

func (user *User) RequestWithCookie (req_url string, method string, data io.Reader) (*http.Response, error) {
	transport := http.Transport{Dial: dialTimeout}
	client := &http.Client{Transport: &transport}
	req, _ := http.NewRequest(method, req_url, data)
	cookie := strings.Join([] string {".SPBForms=", user.spbforms, ";SBVerifyCode=", user.sbvcode}, "")
	req.Header.Set("Cookie", cookie)
	res, err := client.Do(req)
	return res, err
}

func (user *User) Footprinting (req_url string) ([] string, error) {
	response, err := user.RequestWithCookie(req_url, "GET", nil)
	links := make([] string, 0, 10) 
	if (err != nil) {
		panic(err)
	} else{
		defer response.Body.Close()
		if body, err := goquery.NewDocumentFromReader(response.Body); err != nil {
			fmt.Printf("error: %s\n", err)
		} else {
			body.Find("h5").Each(func (i int, s *goquery.Selection) {
				if link, exists := s.Find("a").Attr("href"); exists {
					links = append(links, link)
				}
			})
		}
	} 
	return links, nil
} 

func (user *User) Etl (links [] string) {
	mscnt_regexp := regexp.MustCompile(`(\d+)人参加`)
	date_regexp := regexp.MustCompile(`0?(\d+)月0?(\d+)日`)
	for _, link := range links {
		go func (u User, link string) {
			fmt.Println("Etl <-", link)
			response, err := u.RequestWithCookie(link, "GET", nil)
			if (err != nil) {
				fmt.Println(err)
			} else {
				defer response.Body.Close()
				if rawbody, err := goquery.NewDocumentFromReader(response.Body); err != nil {
					fmt.Printf("error: %s\n", err)
				} else {
					var mscnt int
					var acdate time.Time
					body := rawbody.Find("div[class='tn-box-content tn-widget-content tn-corner-all']")
					subject := rawbody.Find("h1[class='tn-helper-reset tn-text-heading']").Text()
					body.Find("span[class='tn-action']").Find("a").Each(func (i int, s *goquery.Selection) {
						if mscnt_content := mscnt_regexp.FindStringSubmatch(s.Text()); len(mscnt_content) > 1 {
							if cnt, err := strconv.Atoi(mscnt_content[1]); err != nil {
								panic(err)
							} else {
								mscnt = cnt
							}
						}
					})
					if datext := body.Find("span[class='tn-date']").Text(); datext != "" {
						ad, _ := time.Parse("2006年01月02日", "2014年" + date_regexp.FindStringSubmatch(datext)[0])
						acdate = ad
					}
					robbery_body := body.Find("span[class='tn-icon-join tn-icon']").Next()
					robbery_text := robbery_body.Text()
					robbery_addr, _ := robbery_body.Attr("href")
					if strings.Contains(robbery_text, "我要报名") {
						form_response, _ := u.RequestWithCookie(domain + robbery_addr, "GET", nil)
						form_body, _ := goquery.NewDocumentFromReader(form_response.Body)
						if form_addr, form_exists := form_body.Find("form").Attr("action"); form_exists {
							activitie := Activity{subject, acdate, acdate.Weekday(), mscnt, domain + form_addr}
							fmt.Println("Activitys <-", activitie)
							activities <- activitie
						}
					}
				}
			} 
		}(*user, link)
	}
}

func (user *User) Robbery (activity Activity, try_times int) bool {
	if try_times > 3 {
		fmt.Println("已重试3次，报名失败")
	} else if activity.ppcnt >= 25 {
		fmt.Println("<", activity.subject, ">", "活动已满员")
	} else if user.wanna[activity.weekday - 1] == 1 {
		fmt.Printf("第%v次尝试......\n", try_times)
		data := make(url.Values)
		data.Set("body", "spider man")
		data.Set("email", user.email)
		_, err := user.RequestWithCookie(activity.form_addr, "POST", strings.NewReader(data.Encode()))
		if err != nil {
			user.Robbery(activity, try_times + 1)
		} else {
			fmt.Println("<", activity.subject, ">", "报名成功")
			return true
		}
	}
	return false
}

func letusgo() (bool, error) {
	defer func () {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	rint := randint(0, len(users))
	links, err := users[rint].Footprinting(domain + "/c/musle/whatEvents.aspx")
	if err != nil {
		fmt.Println("Footprinting", err)
	}
	users[rint].Etl(links)
	fmt.Println("Etl Running...")
	var ifbreak int
	for ifbreak == 0 {
		select {
    		case <- time.After(time.Second * 3):
       			ifbreak = 1
       			fmt.Println("read channel timeout")
        		break
    		case activity := <- activities:
    			for _, user := range users {
    				wg.Add(1)
    				go func (u User, a Activity) {
    					defer wg.Done()
    					u.Robbery(a, 1)
    				}(user, activity)
    			}
		}
	}
	wg.Wait()
	return true, nil
}

func main() {
	for {
		succ, err := letusgo()
		fmt.Println(succ, err)
	}
}