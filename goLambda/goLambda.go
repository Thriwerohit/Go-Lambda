package golambda

import (
	"context"
	"fmt"
	"log"
	"math"
	"ruleEngine/httpClient"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

func Handler() error {
	// get all events across projectId
	var eventResponse ParseEventResponse
	var userResponse ParseUsers
	//var resp UpdateResponse
	var coin ParseCoins

	// check coin expiry date
	body := strings.NewReader(`{}`)
	_, errCoins := httpClient.PostParseClient("GET", "https://dev.ext-api.thriwe.com/parse/classes/coins", body, &coin)
	if errCoins != nil {
		return errCoins
	}
	for i := 0; i < len(coin.Results); i++ {
		itr := coin.Results[i]
		for j := 0; j < len(itr.Coins); j++ {
			// check for current date and coin exp date
			if itr.Coins[j].ExpiryDate.Day() == time.Now().Day() && itr.Coins[j].ExpiryDate.Month() == time.Now().Month() && itr.Coins[j].ExpiryDate.Year() == time.Now().Year() {

				resp, errUserTracker := UserTracker(itr.UserID)
				if errUserTracker != nil {
					log.Println(errUserTracker.Error())
					return errUserTracker
				}
				if !resp {
					continue
				}

				reqId := RequestIdGenerator()
				errTracker := TrackerSub(itr.ProjectID, itr.UserID, true, int(itr.Coins[j].Amount), "expire", reqId)
				if errTracker != nil {
					fmt.Printf("error occurred while tracking: %v", errTracker)
				}
				err := sendMessageForSubtract(itr.ProjectID, itr.UserID, true, int(itr.Coins[j].Amount), "expire", reqId)
				if err != nil {
					log.Println(err)
					return err
				}
			}
		}
	}

	_, errEvent := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/classes/events", body, &eventResponse)
	if errEvent != nil {
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
				_, errUser := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/users", body, &userResponse)
				if errUser != nil {
					return errUser
				}
				for i := 0; i < len(userResponse.Results); i++ {
					userId := userResponse.Results[i].ObjectID

					date := eventResponse.Results[j].CoinExpiry

					resp, errUserTracker := UserTracker(userId)
					if errUserTracker != nil {
						log.Println(errUserTracker.Error())
						return errUserTracker
					}
					if !resp {
						continue
					}
					reqId := RequestIdGenerator()
					errTracker := Tracker(int(eventResponse.Results[j].CoinAmount), date, userId, eventResponse.Results[j].ProjectID, eventResponse.Results[j].GlobalEventDetails.RuleName, eventResponse.Results[j].ObjectID, reqId)
					if errTracker != nil {
						fmt.Printf("error occurred while tracking: %v", errTracker)
					}
					err := sendMessage(int(eventResponse.Results[j].CoinAmount), date, userId, eventResponse.Results[j].ProjectID, eventResponse.Results[j].GlobalEventDetails.RuleName, eventResponse.Results[j].ObjectID, reqId)
					if err != nil {
						log.Println(err)
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

			_, errUser := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/users", body, &userResponse)
			if errUser != nil {
				return errUser
			}

			for i := 0; i < len(userResponse.Results); i++ {

				dateOfBirth, errTime := time.Parse("2006-01-02", userResponse.Results[i].DateOfBirth)

				if errTime != nil {
					return errTime
				}

				if dateOfBirth.Day() == time.Now().Day() && dateOfBirth.Month() == time.Now().Month() {
					date := eventResponse.Results[j].CoinExpiry
					userId := userResponse.Results[i].ObjectID

					resp, errUserTracker := UserTracker(userId)
					if errUserTracker != nil {
						log.Println(errUserTracker.Error())
						return errUserTracker
					}
					if !resp {
						continue
					}
					reqId := RequestIdGenerator()
					errTracker := Tracker(int(eventResponse.Results[j].CoinAmount), date, userId, eventResponse.Results[j].ProjectID, eventResponse.Results[j].GlobalEventDetails.RuleName, eventResponse.Results[j].ObjectID, reqId)
					if errTracker != nil {
						fmt.Printf("error occurred while tracking: %v", errTracker)
					}
					err := sendMessage(int(eventResponse.Results[j].CoinAmount), date, userId, eventResponse.Results[j].ProjectID, eventResponse.Results[j].GlobalEventDetails.RuleName, eventResponse.Results[j].ObjectID, reqId)
					if err != nil {
						log.Println(err)
						return err
					}
				}
			}
		}
	}

	if int(time.Now().Local().Weekday()) == 2 {
		projects, err := Projects("Weekly")
		if err != nil {
			return err
		}
		if len(projects.Results) == 0 {
			fmt.Println("No weekly projects found")
		}
		prevProjectId := ""
		for i := 0; i < len(projects.Results); i++ {
			if projects.Results[i].ProjectID != prevProjectId {
				errAddCoin := AddCoin(projects.Results[i].ProjectID)
				if errAddCoin != nil {
					return errAddCoin
				}
				prevProjectId = projects.Results[i].ProjectID
			}
		}
	}
	if int(time.Now().Local().Day()) == 1 {
		projects, err := Projects("Monthly")
		if err != nil {
			return err
		}
		if len(projects.Results) == 0 {
			fmt.Println("No weekly projects found")
		}
		prevProjectId := ""
		for i := 0; i < len(projects.Results); i++ {
			if projects.Results[i].ProjectID != prevProjectId {
				errAddCoin := AddCoin(projects.Results[i].ProjectID)
				if errAddCoin != nil {
					return errAddCoin
				}
				prevProjectId = projects.Results[i].ProjectID
			}
		}
	}

	return nil
}

func Projects(repeat string) (*ParseRules, error) {
	var rules ParseRules
	body := strings.NewReader(`{
		"where":{
			"repeatDetail.unit":"` + repeat + `"
		}
	}`)
	_, errProjects := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/classes/rules", body, &rules)
	if errProjects != nil {
		return nil, errProjects
	}
	return &rules, nil
}

func AddCoin(projectId string) error {
	var ruleResponse ParseRules
	var UserResponse ParseUsers
	//var resp UpdateResponse
	bodyRule := strings.NewReader(`{
		"where": {
			"projectId":"` + projectId + `"
		},
		"order":"-expendatureAmount"
	}`)
	_, errProject := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/classes/rules", bodyRule, &ruleResponse)

	if errProject != nil {
		log.Println(errProject.Error())
		return errProject
	}
	// get all user across PROJECTID
	body := strings.NewReader(`{
		"where":{
			"projectId":"` + projectId + `"
		}
	}`)
	_, errUser := httpClient.ParseClient("GET", "https://dev-fab-api-gateway.thriwe.com/parse/users", body, &UserResponse)
	if errUser != nil {
		return errUser
	}

	for i := 0; i < len(UserResponse.Results); i++ {
		// id 1 denotes rule on expendature basis with conversionRate
		coins := 0.0
		var expiry time.Time
		var reason string
		var ruleId string
		for j := 0; j < len(ruleResponse.Results); j++ {
			if ruleResponse.Results[j].RuleExpiryDate.Before(time.Now()) || !ruleResponse.Results[j].IsActive {
				continue
			}

			if ruleResponse.Results[j].GlobalRuleDetails.ID == 1 {

				if UserResponse.Results[i].Expendature > ruleResponse.Results[0].ExpendatureAmount {
					coins = float64(ruleResponse.Results[0].MaxCoin)
					expiry = ruleResponse.Results[j].CoinExpiry
					reason = ruleResponse.Results[j].GlobalRuleDetails.RuleName
					ruleId = ruleResponse.Results[j].ObjectID
				} else {
					expen := float64(UserResponse.Results[i].Expendature)
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
				expen := float64(UserResponse.Results[i].Expendature)
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

			resp, errUserTracker := UserTracker(UserResponse.Results[i].ObjectID)
			if errUserTracker != nil {
				log.Println(errUserTracker.Error())
				return errUserTracker
			}
			if !resp {
				continue
			}
			reqId := RequestIdGenerator()
			errTracker := Tracker(int(coins), expiry, UserResponse.Results[i].ObjectID, projectId, reason, ruleId, reqId)
			if errTracker != nil {
				fmt.Printf("error occurred while tracking: %v", errTracker)
			}
			errMsg := sendMessage(int(coins), expiry, UserResponse.Results[i].ObjectID, projectId, reason, ruleId, reqId)
			if errMsg != nil {
				fmt.Printf("error while sending message: %v", errMsg)
				return errMsg
			}
		}
	}

	return nil
}

func sendMessage(coins int, expiry time.Time, userId string, projectId string, reason string, ruleId, reqId string) error {

	cfg, errLoadDefaultConfig := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("ap-south-1"))
	if errLoadDefaultConfig != nil {
		log.Fatal(errLoadDefaultConfig)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	sqsQueueName := "Test"
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &sqsQueueName,
	}
	result, errGetUrl := GetQueueURL(context.TODO(), sqsClient, gQInput)
	if errGetUrl != nil {
		fmt.Printf("error getting queue url: %v", errGetUrl)
	}

	urlRes := result.QueueUrl

	fmt.Println("Got Queue URL: " + *urlRes)

	queueBody := `{
		"coin":` + fmt.Sprint(coins) + `,
		"expiryDate":"` + expiry.Format("2006-01-02T15:04:05Z") + `",
		"projectId":"` + projectId + `",
		"reason":"` + reason + `",
		"ruleId":"` + ruleId + `",
		"userId":"` + userId + `",
		"typeId":` + fmt.Sprint(1) + `,
		"requestId":"` + reqId + `"
	}`
	sMInput := &sqs.SendMessageInput{
		DelaySeconds: 1,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Title": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Mail"),
			},
			"Author": {
				DataType:    aws.String("String"),
				StringValue: aws.String("ThriweComms library"),
			}},
		MessageBody: aws.String(queueBody),
		QueueUrl:    urlRes,
	}
	resp, err := SendMsg(context.TODO(), sqsClient, sMInput)
	if err != nil {
		fmt.Printf("Got an error while trying to send message to queue: %v", err)
		return err
	}
	fmt.Println("queue response", resp)
	return nil
}

type SQSSendMessageAPI interface {
	GetQueueUrl(ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)

	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

func GetQueueURL(c context.Context, api SQSSendMessageAPI, input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return api.GetQueueUrl(c, input)
}

func SendMsg(c context.Context, api SQSSendMessageAPI, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return api.SendMessage(c, input)
}

func sendMessageForSubtract(projectId string, userId string, isexpire bool, amount int, reason, reqId string) error {
	cfg, errLoadDefaultConfig := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("ap-south-1"))
	if errLoadDefaultConfig != nil {
		log.Fatal(errLoadDefaultConfig)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	sqsQueueName := "Test"
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &sqsQueueName,
	}
	result, errGetUrl := GetQueueURL(context.TODO(), sqsClient, gQInput)
	if errGetUrl != nil {
		fmt.Printf("error getting queue url: %v", errGetUrl)
	}

	urlRes := result.QueueUrl

	fmt.Println("Got Queue URL: " + *urlRes)

	queueBody := `{
		"projectId":"` + projectId + `",
		"amount":` + fmt.Sprint(amount) + `,
		"userId":"` + userId + `",
		"isCoinsExpireReason":` + fmt.Sprint(isexpire) + `,
		"reason":"` + reason + `",
		"typeId":` + fmt.Sprint(2) + `

	}`
	sMInput := &sqs.SendMessageInput{
		DelaySeconds: 1,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Title": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Mail"),
			},
			"Author": {
				DataType:    aws.String("String"),
				StringValue: aws.String("ThriweComms library"),
			}},
		MessageBody: aws.String(queueBody),
		QueueUrl:    urlRes,
	}
	resp, err := SendMsg(context.TODO(), sqsClient, sMInput)
	if err != nil {
		fmt.Printf("Got an error while trying to send message to queue: %v", err)
		return err
	}
	fmt.Println("queue response", resp)
	return nil
}

func RequestIdGenerator() string {
	return random.String(32)
}

func Tracker(coin int, expiry time.Time, userId, projectId, reason, rulId, reqId string) error {
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
		"trackerCreatedAt":"` + time.Now().Local().Format("2006-01-02") + `"
	}`)
	_, err := httpClient.ParseClient("POST", "https://dev-fab-api-gateway.thriwe.com/parse/classes/trackerCoin", body, &resp)
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
