package main

import (
	"fmt"
	"time"
	"os"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type voteBody struct {
	Location string
	Score int
	Apikey string
}

type voteItem struct {
	YearWeekLocation string //composite key
	Year int
	Week int
	Location string
	ScoreSum int
	VoteCount int
}


func requestHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	sess := session.Must(session.NewSession())
	svc  := dynamodb.New(sess)

	now := time.Now()
	year, week := now.ISOWeek()
	weekday := now.Weekday()

	if (weekday != 5 || now.Hour() < 1 || now.Hour() > 3) && os.Getenv("test") != "true" {
		return events.APIGatewayProxyResponse{StatusCode: 403}, nil
	}

	var vb voteBody
	err := json.Unmarshal([]byte(request.Body), &vb)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal json body: %v", err))
	}

	if vb.Score < 1 || vb.Score > 4 {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	compositeKey := fmt.Sprintf("%d-%d-%s", year, week, vb.Location)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("votes"),
		Key: map[string]*dynamodb.AttributeValue{
			"YearWeekLocation": {
					S: aws.String(compositeKey),
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to get item from DB: %v", err))
	}

	item := voteItem{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal DB item: %v", err))
	}

	if item.YearWeekLocation == "" {
		// no record already exist, create
		item.YearWeekLocation = compositeKey
		item.ScoreSum = vb.Score
		item.Location = vb.Location
		item.Week = week
		item.Year = year
		item.VoteCount = 1
	} else {
		// record already exist, update
		item.ScoreSum = item.ScoreSum + vb.Score
		item.VoteCount = item.VoteCount + 1
	}

	mm, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
        panic(fmt.Sprintf("failed to DynamoDB marshal Record, %v", err))
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("votes"),
		Item: mm,
	})


	return events.APIGatewayProxyResponse{Body: string("{\"Status\": \"OK\"}"), StatusCode: 200}, nil


}

func main() {
	lambda.Start(requestHandler)
}
