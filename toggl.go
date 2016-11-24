package toggl

import (
	"encoding/json"
	"fmt"
	"github.com/carlescere/scheduler"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type User struct {
	Id       int `json:"user_id"`
	Duration int `json:"duration"`
}

type Activity struct {
	UserId      int    `json:"user_id"`
	ProjectId   int    `json:"project_id"`
	Duration    int64  `json:"duration"`
	Description string `json:"description"`
	Stop        string `json:"stop"`
	TId         int    `json:"tid"`
}

func (a *Activity) hasJustStarted(interval int64) bool {
	return time.Now().Unix()+a.Duration < interval
}

func (a *Activity) hasJustFinished(interval int64) bool {
	if a.Stop == "" {
		return false
	}
	stop, err := time.Parse(time.RFC3339, a.Stop)
	if err != nil {
		log.Println("ERROR[hasJustFinished]: ", err)
		return false
	}
	return time.Now().Unix()-stop.Unix() < interval
}

type Dashboard struct {
	MostActiveUsers []User     `json:"most_active_user"`
	Activities      []Activity `json:"activity"`
}

func (d *Dashboard) startedActivities(interval int64) []Activity {
	res := make([]Activity, 0)
	for _, a := range d.Activities {
		if a.hasJustStarted(interval) {
			res = append(res, a)
		}
	}
	return res
}

func (d *Dashboard) finishedActivities(interval int64) []Activity {
	res := make([]Activity, 0)
	for _, a := range d.Activities {
		if a.hasJustFinished(interval) {
			res = append(res, a)
		}
	}
	return res
}

func NewHook(interval int64, id int, token string, onStart func(*Activity), onStop func(*Activity), onError func(error)) error {
	f := func() {
		log.Println("toggl-hook started")
		dashboard, err := fetchDashboard(id, token)
		if err != nil {
			onError(err)
		}
		for _, a := range dashboard.finishedActivities(interval) {
			onStop(&a)
		}
		for _, a := range dashboard.startedActivities(interval) {
			onStart(&a)
		}
		log.Println("toggl-hook finished")
	}
	_, err := scheduler.Every(int(interval)).Seconds().Run(f)
	return err
}

func fetchDashboard(id int, token string) (*Dashboard, error) {
	c := &http.Client{}
	url := fmt.Sprintf("https://www.toggl.com/api/v8/dashboard/%d", id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(token, "api_token")

	log.Printf("%v %v %v\n", req.Proto, req.Method, req.URL)

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("%v %v\n", resp.Proto, resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	d := &Dashboard{}
	err = json.Unmarshal(body, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}
