package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/labstack/gommon/random"
)

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

func RequestIdGenerator() string {
	return random.String(32)
}

func SendMessage(coins int, expiry time.Time, userId string, projectId string, reason string, ruleId, reqId string, id int) error {
	cfg, errLoadDefaultConfig := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(os.Getenv("region")))
	if errLoadDefaultConfig != nil {
		log.Fatal(errLoadDefaultConfig)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	sqsQueueName := os.Getenv("queueName")
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &sqsQueueName,
	}
	result, errGetUrl := GetQueueURL(context.TODO(), sqsClient, gQInput)
	if errGetUrl != nil {
		fmt.Printf("error getting queue url: %v", errGetUrl.Error())
		return errGetUrl
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
		"requestId":"` + reqId + `",
		"eventId":` + fmt.Sprint(id) + `
	}`
	sMInput := &sqs.SendMessageInput{
		DelaySeconds: 1,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Title": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Points"),
			},
			"Author": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Points Module"),
			}},
		MessageBody: aws.String(queueBody),
		QueueUrl:    urlRes,
	}
	resp, err := SendMsg(context.TODO(), sqsClient, sMInput)
	if err != nil {
		fmt.Printf("Got an error while trying to send message to queue: %v", err.Error())
		return err
	}
	fmt.Println("queue response", resp)
	return nil
}

func SendMessageForSubtract(projectId string, userId string, isexpire bool, amount int, reason, reqId string, id int) error {
	cfg, errLoadDefaultConfig := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(os.Getenv("region")))
	if errLoadDefaultConfig != nil {
		log.Fatal(errLoadDefaultConfig)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	sqsQueueName := os.Getenv("queueName")
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &sqsQueueName,
	}
	result, errGetUrl := GetQueueURL(context.TODO(), sqsClient, gQInput)
	if errGetUrl != nil {
		fmt.Printf("error getting queue url: %v", errGetUrl.Error())
		return errGetUrl
	}

	urlRes := result.QueueUrl

	fmt.Println("Got Queue URL: " + *urlRes)

	queueBody := `{
		"projectId":"` + projectId + `",
		"amount":` + fmt.Sprint(amount) + `,
		"userId":"` + userId + `",
		"isCoinsExpireReason":` + fmt.Sprint(isexpire) + `,
		"reason":"` + reason + `",
		"typeId":` + fmt.Sprint(2) + `,
		"requestId":"` + reqId + `",
		"eventId":` + fmt.Sprint(id) + `
	}`
	sMInput := &sqs.SendMessageInput{
		DelaySeconds: 1,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Title": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Points"),
			},
			"Author": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Points Module"),
			}},
		MessageBody: aws.String(queueBody),
		QueueUrl:    urlRes,
	}
	resp, err := SendMsg(context.TODO(), sqsClient, sMInput)
	if err != nil {
		fmt.Printf("Got an error while trying to send message to queue: %v", err.Error())
		return err
	}
	fmt.Println("queue response", resp)
	return nil
}
