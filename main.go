package main

import (
	"encoding/json"
	"ruleEngine/httpClient"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	// e := echo.New()
	// e.POST("/test",Handler)
	// e.Logger.Fatal(e.Start(":8000"))
	lambda.Start(Handler)

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

type UpdateResponse []struct {
	Success struct {
		UpdatedAt time.Time `json:"updatedAt"`
	} `json:"success"`
}

func Handler() error {
	// get all events across projectId
	var eventResponse ParseEventResponse
	var userResponse ParseUsers
	var addCoin AddCoinRequest
	var resp UpdateResponse
	_, errEvent := httpClient.ParseClient("GET", "http://localhost:1337/parse/classes/events", nil, &eventResponse)
	if errEvent != nil {
		return errEvent
	}
	// get coins
	_, errUser := httpClient.ParseClient("GET", "http://localhost:1337/parse/classes/users", nil, &userResponse)
	if errEvent != nil {
		return errUser
	}
	for i := 0; i < len(userResponse.Results); i++ {
		for j := 0; j < len(eventResponse.Results); j++ {
			if eventResponse.Results[j].GlobalEventDetails.ID == 1 && eventResponse.Results[j].ProjectID == userResponse.Results[i].ProjectId {
				if eventResponse.Results[j].EventDate.Day() == time.Now().Day() && eventResponse.Results[j].EventDate.Month() == time.Now().Month() {
					addCoin.UserId = userResponse.Results[i].ObjectID
					addCoin.ProjectId = userResponse.Results[i].ProjectId
					addCoin.ExpiryDate = eventResponse.Results[j].CoinExpiry
					addCoin.Coin = float64(eventResponse.Results[j].CoinAmount)
					b, err := json.Marshal(addCoin)
					if err!= nil {
                        return err
                    }
					body := strings.NewReader(string(b))
					_,errCoins := httpClient.ParseClient("PUT","http://localhost:8080/addCoin",body,&resp)
					if errCoins!= nil {
                        return errCoins
                    }
				}
			}
		}
	}
	return nil
}
