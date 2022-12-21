package main

import (
	"fmt"
	"ruleEngine/httpClient"
	"strings"
	"time"
	//"github.com/labstack/echo"
	//"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	//e := echo.New()
	//e.POST("/test", Handler)
	//e.Logger.Fatal(e.Start(":8000"))
	//lambda.Start(Handler)
	Handler()
}

type Event struct {
	ObjectID           string `json:"objectId"`
	ProjectID          string `json:"projectId"`
	GlobalEventDetails struct {
		ID       int    `json:"id"`
		RuleName string `json:"ruleName"`
	} `json:"globalEventDetails"`
	IsRepeat     bool `json:"isRepeat"`
	RepeatDetail struct {
		Value int    `json:"value"`
		Unit  string `json:"unit"`
	} `json:"repeatDetail"`
	CoinAmount     int       `json:"coinAmount"`
	CoinExpiry     time.Time `json:"coinExpiry"`
	IsActive       bool      `json:"isActive"`
	RuleExpiryDate time.Time `json:"ruleExpiryDate"`
	EventDate      time.Time `json:"eventDate"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ParseEventResponse struct {
	Results []struct {
		Event
	} `json:"results"`
}
type Users struct {
	ObjectID     string `json:"objectId"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Username     string `json:"username"`
	MobileNumber string `json:"mobileNumber"`
	CountryCode  string `json:"countryCode"`
	Expendature  int    `json:"expendature"`
	ProjectId    string `json:"projectId"`
}

type ParseUsers struct {
	Results []struct {
		Users
	} `json:"results"`
}
type AddCoinRequest struct {
	ProjectId  string    `json:"projectId"`
	UserId     string    `json:"userId"`
	Coin       float64   `json:"coin"`
	ExpiryDate time.Time `json:"expiryDate"`
}
type Coin struct {
	ObjectID  string    `json:"objectId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ProjectID string    `json:"projectId"`
	UserID    string    `json:"userId"`
	Coins     []struct {
		Amount     float64   `json:"amount"`
		ExpiryDate time.Time `json:"expiryDate"`
	} `json:"coins"`
	TotalCoins int `json:"totalCoins"`
}

type ParseCoins struct {
	Results []struct {
		Coin
	} `json:"results"`
}
type UpdateResponse []struct {
	Success struct {
		UpdatedAt time.Time `json:"updatedAt"`
	} `json:"success"`
}

type SubtractCoinRequest struct {
	ProjectId           string  `json:"projectId"`
	UserId              string  `json:"userId"`
	IsCoinsExpireReason bool    `json:"isCoinsExpireReason"`
	Amount              float64 `json:"amount"`
	Reason              string  `json:"reason"`
}

func Handler() error {
	// get all events across projectId
	var eventResponse ParseEventResponse
	var userResponse ParseUsers
	//var addCoin AddCoinRequest
	var resp UpdateResponse
	var coin ParseCoins

	// check coin expiry date
	body := strings.NewReader(`{}`)
	_, errCoins := httpClient.ParseClient("GET", "http://localhost:8080/parse/classes/coins",body, &coin)
	if errCoins != nil {
		return errCoins
	}
	for i := 0; i < len(coin.Results); i++ {
		itr := coin.Results[i]
		for j := 0; j < len(itr.Coins); j++ {
			// check for current date and coin exp date
			if itr.Coins[j].ExpiryDate.Day() == time.Now().Day() && itr.Coins[j].ExpiryDate.Month() == time.Now().Month() && itr.Coins[j].ExpiryDate.Year() == time.Now().Year() {
				coinSubtractBody := strings.NewReader(`{
					"projectId":"` + itr.ProjectID + `",
					"userId":"` + itr.UserID + `",
					"isCoinsExpireReason":` + fmt.Sprint(true) + `,
					"amount":` + fmt.Sprint(itr.Coins[j].Amount) + `,
					"reason": ""
				 }`)
				_, errSubtractCoins := httpClient.ParseClient("PUT", "http://localhost:8083/subtractCoins", coinSubtractBody, &resp)
				if errSubtractCoins != nil {
					return errSubtractCoins
				}
			}
		}
	}

	_, errEvent := httpClient.ParseClient("GET", "http://localhost:1337/parse/classes/events", body, &eventResponse)
	if errEvent != nil {
		return errEvent
	}
	// get coins
	_, errUser := httpClient.ParseClient("GET", "http://localhost:1337/parse/classes/users", body, &userResponse)
	if errEvent != nil {
		return errUser
	}
	for i := 0; i < len(userResponse.Results); i++ {
		for j := 0; j < len(eventResponse.Results); j++ {
			if eventResponse.Results[j].GlobalEventDetails.ID == 1 && eventResponse.Results[j].ProjectID == userResponse.Results[i].ProjectId {
				if eventResponse.Results[j].EventDate.Day() == time.Now().Day() && eventResponse.Results[j].EventDate.Month() == time.Now().Month() {
					date := eventResponse.Results[j].EventDate
					body := strings.NewReader(`{
						"projectId":"` + eventResponse.Results[j].ProjectID + `",
                        "userId":"` + userResponse.Results[i].ObjectID + `",
                        "coin":` + fmt.Sprint(eventResponse.Results[j].CoinAmount) + `,
						"expiryDate":"` + date.Format("2006-01-02T15:04:05Z") + `"
					}`)
					_, errCoins := httpClient.ParseClient("PUT", "http://localhost:8083/addCoins", body, &resp)
					if errCoins != nil {
						return errCoins
					}
				}
			}
		}
	}
	return nil
}