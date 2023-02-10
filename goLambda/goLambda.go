package golambda

import (
	"fmt"
	"log"
	"math"
	aws "ruleEngine/aws"
	"ruleEngine/httpClient"
	tracker "ruleEngine/tracker"
	"strings"
	"time"

	"github.com/labstack/gommon/random"
)

type Event struct {
	ObjectID           string `json:"objectId"`
	ProjectID          string `json:"projectId"`
	GlobalEventDetails struct {
		ID       int    `json:"id"`
		RuleName string `json:"ruleName"`
	} `json:"globalEventDetails"`
	Name         string `json:"name"`
	IsRepeat     bool   `json:"isRepeat"`
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
	UserDetails    []struct {
		UserID      string    `json:"userId"`
		DateOfBirth time.Time `json:"dateOfBirth"`
	} `json:"userDetails"`
}

type ParseEventResponse struct {
	Results []struct {
		Event
	} `json:"results"`
}
type Users struct {
	ObjectID     string `json:"objectId"`
	ProjectId    string `json:"projectId"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	LastName     string `json:"lastName"`
	EmailID      string `json:"emailId"`
	MobileNumber string `json:"mobileNumber"`
	DateOfBirth  string `json:"dateOfBirth"`
	Expendature  int    `json:"expendature"`
}

type ParseUsers struct {
	Results []struct {
		Users
	} `json:"results"`
}

type Expendature struct {
	ObjectID           string `json:"objectId"`
	UserID             string `json:"userId"`
	Expendature        int    `json:"expendature"`
	MonthlyExpendature int    `json:"monthlyExpendature"`
	WeeklyExpendature  int    `json:"weeklyExpendature"`
	Logs               []struct {
		Coins   int    `json:"coins"`
		AddedAt string `json:"addedAt"`
	} `json:"logs"`
}
type ParseExpedature struct {
	Results []struct {
		Expendature
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

type ExpendatureRule struct {
	ObjectID          string `json:"objectId"`
	ProjectID         string `json:"projectId"`
	GlobalRuleDetails struct {
		ID       int    `json:"id"`
		RuleName string `json:"ruleName"`
	} `json:"globalRuleDetails"`
	IsRepeat     bool   `json:"isRepeat"`
	Name         string `json:"name"`
	RepeatDetail struct {
		Value int    `json:"value"`
		Unit  string `json:"unit"`
	} `json:"repeatDetail"`
	CoinAmount        int       `json:"coinAmount"`
	CoinExpiry        time.Time `json:"coinExpiry"`
	IsActive          bool      `json:"isActive"`
	RuleExpiryDate    time.Time `json:"ruleExpiryDate"`
	MaxCoin           int       `json:"MaxCoin"`
	ExpendatureAmount int       `json:"expendatureAmount"`
	ConversionRate    float64   `json:"conversionRate"`
}
type ParseRules struct {
	Results []struct {
		ExpendatureRule
	} `json:"results"`
}

type ParseCreate struct {
	ObjectID  string    `json:"objectId"`
	CreatedAt time.Time `json:"createdAt"`
}

func RequestIdGenerator() string {
	return random.String(32)
}

func Handler() error {
	// get all events across projectId
	var eventResponse ParseEventResponse
	var userResponse ParseUsers
	//var resp UpdateResponse
	var coin ParseCoins

	// check coin expiry date
	body := strings.NewReader(`{}`)
	_, errCoins := httpClient.PostParseClient("GET", "classes/coins", body, &coin)
	if errCoins != nil {
		log.Println("error while getting coin", errCoins)
		return errCoins
	}
	for i := 0; i < len(coin.Results); i++ {
		itr := coin.Results[i]
		for j := 0; j < len(itr.Coins); j++ {
			// check for current date and coin exp date
			if itr.Coins[j].ExpiryDate.Day() == time.Now().Day() && itr.Coins[j].ExpiryDate.Month() == time.Now().Month() && itr.Coins[j].ExpiryDate.Year() == time.Now().Year() {

				resp, errUserTracker := tracker.UserTracker(itr.UserID, 1)
				if errUserTracker != nil {
					log.Println("tracker error", errUserTracker.Error())
					return errUserTracker
				}
				if !resp {
					continue
				}

				reqId := RequestIdGenerator()
				errTracker := tracker.TrackerSub(itr.ProjectID, itr.UserID, true, int(itr.Coins[j].Amount), "expire", reqId, 1)
				if errTracker != nil {
					fmt.Printf("error occurred while tracking: %v", errTracker)
					return errTracker
				}
				err := aws.SendMessageForSubtract(itr.ProjectID, itr.UserID, true, int(itr.Coins[j].Amount), "expire", reqId, 1)
				if err != nil {
					log.Println("error while sending msg", err.Error())
					return err
				}
			}
		}
	}

	_, errEvent := httpClient.ParseClient("GET", "classes/events", body, &eventResponse)
	if errEvent != nil {
		log.Println("error on getting event response", errEvent)
		return errEvent
	}

	for j := 0; j < len(eventResponse.Results); j++ {

		if eventResponse.Results[j].RuleExpiryDate.Before(time.Now()) || !eventResponse.Results[j].IsActive {
			continue
		}

		if eventResponse.Results[j].GlobalEventDetails.ID == 1 {
			if eventResponse.Results[j].EventDate.Day() == time.Now().Day() && eventResponse.Results[j].EventDate.Month() == time.Now().Month() {

				body := strings.NewReader(`{
					"where":{
						"projectId":"` + eventResponse.Results[j].ProjectID + `"
					}
				}`)
				_, errUser := httpClient.ParseClient("GET", "users", body, &userResponse)
				if errUser != nil {
					log.Println("error while getting users in custom event", errUser.Error())
					return errUser
				}
				for i := 0; i < len(userResponse.Results); i++ {
					userId := userResponse.Results[i].ObjectID

					date := eventResponse.Results[j].CoinExpiry

					resp, errUserTracker := tracker.UserTracker(userId, 2)
					if errUserTracker != nil {
						log.Println("tracker error custom event", errUserTracker.Error())
						return errUserTracker
					}
					if !resp {
						continue
					}
					reqId := RequestIdGenerator()
					errTracker := tracker.TrackerAddCoin(int(eventResponse.Results[j].CoinAmount), date, userId, eventResponse.Results[j].ProjectID, eventResponse.Results[j].GlobalEventDetails.RuleName, eventResponse.Results[j].ObjectID, reqId, 2)
					if errTracker != nil {
						fmt.Printf("error occurred while tracking custom event: %v", errTracker.Error())
						return errTracker
					}
					err := aws.SendMessage(int(eventResponse.Results[j].CoinAmount), date, userId, eventResponse.Results[j].ProjectID, eventResponse.Results[j].GlobalEventDetails.RuleName, eventResponse.Results[j].ObjectID, reqId, 2)
					if err != nil {
						log.Println("error while sending msg in custom event", err.Error())
						return err
					}
				}
			}
		} else if eventResponse.Results[j].GlobalEventDetails.ID == 2 {
			body := strings.NewReader(`{
				"where":{
					"projectId":"` + eventResponse.Results[j].ProjectID + `"
				}
			}`)

			_, errUser := httpClient.ParseClient("GET", "users", body, &userResponse)
			if errUser != nil {
				log.Println("error while getting user in birthday event", errUser.Error())
				return errUser
			}

			for i := 0; i < len(userResponse.Results); i++ {

				dateOfBirth, errTime := time.Parse("2006-01-02", userResponse.Results[i].DateOfBirth)

				if errTime != nil {
					log.Println("error in DOB", errTime.Error())
					return errTime
				}

				if dateOfBirth.Day() == time.Now().Day() && dateOfBirth.Month() == time.Now().Month() {
					date := eventResponse.Results[j].CoinExpiry
					userId := userResponse.Results[i].ObjectID

					resp, errUserTracker := tracker.UserTracker(userId, 3)
					if errUserTracker != nil {
						log.Println("error while tracking in birthday", errUserTracker.Error())
						return errUserTracker
					}
					if !resp {
						continue
					}
					reqId := RequestIdGenerator()
					errTracker := tracker.TrackerAddCoin(int(eventResponse.Results[j].CoinAmount), date, userId, eventResponse.Results[j].ProjectID, eventResponse.Results[j].GlobalEventDetails.RuleName, eventResponse.Results[j].ObjectID, reqId, 3)
					if errTracker != nil {
						fmt.Printf("error occurred while tracking in birthday: %v", errTracker.Error())
						return errTracker
					}
					err := aws.SendMessage(int(eventResponse.Results[j].CoinAmount), date, userId, eventResponse.Results[j].ProjectID, eventResponse.Results[j].GlobalEventDetails.RuleName, eventResponse.Results[j].ObjectID, reqId, 3)
					if err != nil {
						log.Println(err)
						return err
					}
				}
			}
		}
	}

	if int(time.Now().Local().Weekday()) == 5 {
		errAddCoin := AddCoin("XlsTfS4VXW", "Weekly")
		if errAddCoin != nil {
			log.Println("error while adding coin", errAddCoin.Error())
			return errAddCoin
		}
		if int(time.Now().Local().Day()) == 17 {

			errAddCoin := AddCoin("XlsTfS4VXW", "Monthly")
			if errAddCoin != nil {
				log.Println("error while adding coin", errAddCoin.Error())
				return errAddCoin

			}
		}
	}
	return nil
}

func AddCoin(projectId, expense string) error {
	var ruleResponse ParseRules
	var UserResponse ParseExpedature
	//var resp UpdateResponse
	bodyRule := strings.NewReader(`{
		"where": {
			"repeatDetail.unit":"` + expense + `"
		},
		"order":"-expendatureAmount"
	}`)
	_, errProject := httpClient.ParseClient("GET", "classes/rules", bodyRule, &ruleResponse)

	if errProject != nil {
		log.Println("error in getting rules in add coin", errProject.Error())
		return errProject
	}
	// get all user across PROJECTID
	body := strings.NewReader(`{
	}`)
	_, errUser := httpClient.ParseClient("GET", "classes/expendatureLogs", body, &UserResponse)
	if errUser != nil {
		log.Println("error in getting rules in add coin", errUser.Error())
		return errUser
	}

	for i := 0; i < len(UserResponse.Results); i++ {
		// id 1 denotes rule on expendature basis with conversionRate
		coins := 0.0
		var expiry time.Time
		var reason string
		var ruleId string
		var exp int
		if expense == "Weekly" {
			exp = UserResponse.Results[i].WeeklyExpendature
		} else if expense == "Monthly" {
			exp = UserResponse.Results[i].MonthlyExpendature
		}
		for j := 0; j < len(ruleResponse.Results); j++ {
			if ruleResponse.Results[j].RuleExpiryDate.Before(time.Now()) || !ruleResponse.Results[j].IsActive {
				continue
			}

			if ruleResponse.Results[j].GlobalRuleDetails.ID == 1 {

				if exp > ruleResponse.Results[0].ExpendatureAmount {
					coins = float64(ruleResponse.Results[0].MaxCoin)
					expiry = ruleResponse.Results[j].CoinExpiry
					reason = ruleResponse.Results[j].GlobalRuleDetails.RuleName
					ruleId = ruleResponse.Results[j].ObjectID
				} else {
					expen := float64(exp)
					conversionRate := ruleResponse.Results[j].ConversionRate
					coinsCoversion := expen * conversionRate
					if coinsCoversion > coins {
						coins = coinsCoversion
						expiry = ruleResponse.Results[j].CoinExpiry
						reason = ruleResponse.Results[j].GlobalRuleDetails.RuleName
						ruleId = ruleResponse.Results[j].ObjectID
					}

				}

			} else if ruleResponse.Results[j].GlobalRuleDetails.ID == 2 {
				expen := float64(exp)
				conversionRate := ruleResponse.Results[j].ConversionRate
				coinsFlat := expen * conversionRate

				if coinsFlat > coins {
					coins = coinsFlat
					expiry = ruleResponse.Results[j].CoinExpiry
					reason = ruleResponse.Results[j].GlobalRuleDetails.RuleName
					ruleId = ruleResponse.Results[j].ObjectID
				}

			}
		}

		if coins != 0.0 {
			coins = math.Floor(coins*100) / 100

			if expense == "Weekly" {
				resp, errUserTracker := tracker.UserTracker(UserResponse.Results[i].UserID, 4)
				if errUserTracker != nil {
					log.Println("error in tracking in add coin", errUserTracker.Error())
					return errUserTracker
				}

				if !resp {
					continue
				}

				reqId := RequestIdGenerator()

				errTracker := tracker.TrackerAddCoin(int(coins), expiry, UserResponse.Results[i].UserID, projectId, reason, ruleId, reqId, 4)
				if errTracker != nil {
					fmt.Printf("error occurred while tracking in add coin: %v", errTracker.Error())
				}

				// errMsg := aws.SendMessage(int(coins), expiry, UserResponse.Results[i].UserID, projectId, reason, ruleId, reqId, 4)
				// if errMsg != nil {
				// 	fmt.Printf("error while sending message in add coin: %v", errMsg.Error())
				// 	return errMsg
				// }

			} else if expense == "Monthly" {
				resp, errUserTracker := tracker.UserTracker(UserResponse.Results[i].UserID, 5)
				if errUserTracker != nil {
					log.Println("error in tracking in add coin", errUserTracker.Error())
					return errUserTracker
				}

				if !resp {
					continue
				}

				reqId := RequestIdGenerator()
				errTracker := tracker.TrackerAddCoin(int(coins), expiry, UserResponse.Results[i].UserID, projectId, reason, ruleId, reqId, 5)
				if errTracker != nil {
					fmt.Printf("error occurred while tracking in add coin: %v", errTracker.Error())
				}

				errMsg := aws.SendMessage(int(coins), expiry, UserResponse.Results[i].UserID, projectId, reason, ruleId, reqId, 5)
				if errMsg != nil {
					fmt.Printf("error while sending message in add coin: %v", errMsg.Error())
					return errMsg
				}
			}

		}
	}

	return nil
}

