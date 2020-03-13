package main

import (
	"fmt"
	"time"
	"strconv"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type kakItem struct {
	Year int
	Week int
	Name string
	Url string
}

type thisweekResponse struct {
	Kaka string
	Url string
}

func requestHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	sess := session.Must(session.NewSession())
	svc  := dynamodb.New(sess)

	year, week := time.Now().ISOWeek()

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("kakor"),
		Key: map[string]*dynamodb.AttributeValue{
			"Year": {
					N: aws.String(strconv.Itoa(year)),
			},
			"Week": {
				N: aws.String(strconv.Itoa(week)),
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to get item from DB: %v", err))
	}

	item := kakItem{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal DB item: %v", err))
	}

	if item.Name == "" || item.Url == "" {
		return events.APIGatewayProxyResponse{StatusCode: 418}, nil
	}

	s, _ := json.Marshal(&thisweekResponse{item.Name, item.Url})

	return events.APIGatewayProxyResponse{Body: string(s), StatusCode: 200}, nil


}

func main() {
	lambda.Start(requestHandler)
}
