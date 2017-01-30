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

// Activities do not have id and do not include time_entry_id also.
// So identify activities by (user_id, description)
type Activity struct {
	UserId      int    `json:"user_id"`
	ProjectId   int    `json:"project_id"`
	Duration    int64  `json:"duration"`
	Description string `json:"description"`
	Stop        string `json:"stop"`
	TId         int    `json:"tid"`
}

func (a *Activity) hasJustStarted(runningActivities map[int]string, interval int64) bool {
	d, didRun := runningActivities[a.UserId]
	return ((!didRun || d != a.Description) &&
		time.Now().Unix()+a.Duration < int64(1.5*float64(interval))) // margin to fill gap
}

func (a *Activity) hasJustFinished(runningActivities map[int]string) bool {
	if a.Stop == "" {
		return false
	}
	d, didRun := runningActivities[a.UserId]
	return didRun && d == a.Description
}

type Dashboard struct {
	client            *http.Client
	request           *http.Request
	runningActivities map[int]string
	MostActiveUsers   []User     `json:"most_active_user"`
	Activities        []Activity `json:"activity"`
}

func NewDashboard(id int, token string) (*Dashboard, error) {
	c := &http.Client{}
	url := fmt.Sprintf("https://www.toggl.com/api/v8/dashboard/%d", id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(token, "api_token")

	return &Dashboard{client: c, request: req, runningActivities: make(map[int]string)}, nil
}

func (d *Dashboard) fetch() error {
	log.Printf("%v %v %v\n", d.request.Proto, d.request.Method, d.request.URL)

	resp, err := d.client.Do(d.request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("%v %v\n", resp.Proto, resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	d.Activities = nil
	err = json.Unmarshal(body, d)
	if err != nil {
		return err
	}

	return nil
}

func (d *Dashboard) latestActivities() []Activity {
	res := make([]Activity, 0)
	checked := map[int]bool{}
	for _, a := range d.Activities {
		if checked[a.UserId] {
			continue
		} else {
			res = append(res, a)
			checked[a.UserId] = true
		}
	}
	return res
}

func (d *Dashboard) startedActivities(interval int64) []Activity {
	res := make([]Activity, 0)
	for _, a := range d.latestActivities() {
		if a.hasJustStarted(d.runningActivities, interval) {
			d.runningActivities[a.UserId] = a.Description
			res = append(res, a)
		}
	}
	return res
}

func (d *Dashboard) finishedActivities() []Activity {
	res := make([]Activity, 0)
	for _, a := range d.latestActivities() {
		if a.hasJustFinished(d.runningActivities) {
			delete(d.runningActivities, a.UserId)
			res = append(res, a)
		}
	}
	return res
}

func NewHook(interval int64, id int, token string, onStart func(*Activity), onStop func(*Activity), onError func(error)) error {
	// There is no endpoint to get workspace wide time entries other than /dashboard/:id
	dashboard, err := NewDashboard(id, token)
	if err != nil {
		return err
	}

	f := func() {
		log.Println("toggl-hook started")
		err = dashboard.fetch()
		if err != nil {
			onError(err)
		}
		for _, a := range dashboard.finishedActivities() {
			onStop(&a)
		}
		for _, a := range dashboard.startedActivities(interval) {
			onStart(&a)
		}
		log.Println("toggl-hook finished")
	}
	_, err = scheduler.Every(int(interval)).Seconds().Run(f)
	return err
}
