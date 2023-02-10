package tracker

import (
	"fmt"
	httpClient "ruleEngine/httpClient"
	"strings"
	"time"
)

type ParseCreate struct {
	ObjectID  string    `json:"objectId"`
	CreatedAt time.Time `json:"createdAt"`
}

func TrackerAddCoin(coin int, expiry time.Time, userId, projectId, reason, rulId, reqId string, id int) error {
	var resp ParseCreate
	body := strings.NewReader(`{
		"requestId":"` + reqId + `",
		"userId":"` + userId + `",
		"projectId":"` + projectId + `",
		"reason":"` + reason + `",
		"coin":` + fmt.Sprint(coin) + `,
		"expiryDate":"` + expiry.Format("2006-01-02T15:04:05Z") + `",
		"status": "initiated",
		"ruleId": "` + rulId + `",
		"typeId": ` + fmt.Sprint(1) + `,
		"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `",
		"eventId":` + fmt.Sprint(id) + `
	}`)

	_, err := httpClient.ParseClient("POST", "classes/trackerCoin", body, &resp)
	if err != nil {
		fmt.Print("error while creating tracker", reqId)
		return err
	}
	return nil
}


func TrackerSub(projectId string, userId string, isExpire bool, coin int, reason, reqId string, id int) error {
	var resp ParseCreate

	body := strings.NewReader(`{
		"requestId":"` + reqId + `",
		"userId":"` + userId + `",
		"projectId":"` + projectId + `",
		"reason":"` + reason + `",
		"coin":` + fmt.Sprint(coin) + `,
		"isCoinsExpireReason":` + fmt.Sprint(isExpire) + `,
		"typeId":` + fmt.Sprint(2) + `,
		"status": "initiated",
		"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `",
		"eventId":` + fmt.Sprint(id) + `
	}`)
	_, err := httpClient.ParseClient("POST", "classes/trackerCoin", body, &resp)
	if err != nil {
		fmt.Print("error while creating tracker", reqId)
		return err
	}
	return nil
}

type RequestTracker struct {
	ProjectId           string    `json:"projectId"`
	UserId              string    `json:"userId"`
	IsCoinsExpireReason bool      `json:"isCoinsExpireReason"`
	Amount              float64   `json:"amount"`
	ObjectID            string    `json:"objectId"`
	Coin                int       `json:"coin"`
	ExpirtDate          time.Time `json:"expirtDate"`
	RuleID              string    `json:"ruleId"`
	Reason              string    `json:"reason"`
	RequestID           string    `json:"requestId"`
	TrackerCreatedAt    string    `json:"trackerCreatedAt"`
	Status              string    `json:"status"`
	TypeId              int       `json:"typeId"`
}

type ParseTracker struct {
	Results []struct {
		RequestTracker
	} `json:"results"`
}
	
func UserTracker(userId string, id int) (bool, error) {
	var tracker ParseTracker
	bodyTracker := strings.NewReader(`{
		"where":{
			"userId":"` + userId + `",
			"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `",
			"eventId":` + fmt.Sprint(id) + `
		}
	}`)
	_, errTracker := httpClient.ParseClient("GET", "classes/trackerCoin", bodyTracker, &tracker)
	if errTracker != nil {
		return false, errTracker
	}
	if len(tracker.Results) == 0 {
		return true, nil
	} else {
		return false, nil
	}

}


