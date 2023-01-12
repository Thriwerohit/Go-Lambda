package tracker

import (
	"fmt"
	"ruleEngine/httpClient"
	"strings"
	"time"
)

type ParseCreate struct {
	ObjectID  string    `json:"objectId"`
	CreatedAt time.Time `json:"createdAt"`
}

func TrackerRuleCoin(coin int, expiry time.Time, userId, projectId, reason, rulId, reqId string, week, month bool) error {
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
		"week":` + fmt.Sprint(week) + `,
		"month":` + fmt.Sprint(month) + `
	}`)

	_, err := httpClient.ParseClient("POST", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerCoin", body, &resp)
	if err != nil {
		fmt.Print("error while creating tracker", reqId)
		return err
	}
	return nil
}

func TrackerEventCoin(coin int, expiry time.Time, userId, projectId, reason, rulId, reqId string, custom, birthday bool) error {
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
		"custom":` + fmt.Sprint(custom) + `,
		"birthday":` + fmt.Sprint(birthday) + `,
		"isEvent":` + fmt.Sprint(true) + `
	}`)
	_, err := httpClient.ParseClient("POST", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerEventCoin", body, &resp)
	if err != nil {
		fmt.Print("error while creating tracker", reqId)
		return err
	}
	return nil
}

func TrackerSub(projectId string, userId string, isExpire bool, coin int, reason, reqId string) error {
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
		"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `"
	}`)
	_, err := httpClient.ParseClient("POST", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerCoin", body, &resp)
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

func UserTrackerWeek(userId, value string) (bool, error) {
	var tracker ParseTracker
	if value == "Weekly" {
		bodyTracker := strings.NewReader(`{
		"where":{
			"userId":"` + userId + `",
			"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `",
			"week":` + fmt.Sprint(true) + `
		}
	}`)
		_, errTracker := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerCoin", bodyTracker, &tracker)
		if errTracker != nil {
			return false, errTracker
		}
	} else if value == "Monthly" {
		bodyTracker := strings.NewReader(`{
			"where":{
				"userId":"` + userId + `",
				"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `",
				"month":` + fmt.Sprint(true) + `
			}
		}`)
		_, errTracker := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerCoin", bodyTracker, &tracker)
		if errTracker != nil {
			return false, errTracker
		}
	}
	if len(tracker.Results) == 0 {
		return true, nil
	} else {
		return false, nil
	}

}
func UserTracker(userId string) (bool, error) {
	var tracker ParseTracker
	bodyTracker := strings.NewReader(`{
		"where":{
			"userId":"` + userId + `",
			"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `"
		}
	}`)
	_, errTracker := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerCoin", bodyTracker, &tracker)
	if errTracker != nil {
		return false, errTracker
	}
	if len(tracker.Results) == 0 {
		return true, nil
	} else {
		return false, nil
	}

}
func UserTrackerCustom(userId string) (bool, error) {
	var tracker ParseTracker
	bodyTracker := strings.NewReader(`{
		"where":{
			"userId":"` + userId + `",
			"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `",
			"custom":` + fmt.Sprint(true) + `
		}
	}`)
	_, errTracker := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerEventCoin", bodyTracker, &tracker)
	if errTracker != nil {
		return false, errTracker
	}
	if len(tracker.Results) == 0 {
		return true, nil
	} else {
		return false, nil
	}

}
func UserTrackerBirthday(userId string) (bool, error) {
	var tracker ParseTracker
	bodyTracker := strings.NewReader(`{
		"where":{
			"userId":"` + userId + `",
			"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `",
			"birthday":` + fmt.Sprint(true) + `
		}
	}`)
	_, errTracker := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerEventCoin", bodyTracker, &tracker)
	if errTracker != nil {
		return false, errTracker
	}
	if len(tracker.Results) == 0 {
		return true, nil
	} else {
		return false, nil
	}

}
