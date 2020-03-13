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

type score struct {
	Location string
	Score float32
}

type kakResponse2 struct {
	Kaka string
	Year int
	Week int
	Scores []score
}

func getScores(year, week int) []score {
	var scores []score
	scores = append(scores, score{"test", 1.2})
	return scores
}

func requestHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	sess := session.Must(session.NewSession())
	svc  := dynamodb.New(sess)


	year := request.PathParameters["year"]
	week := request.PathParameters["week"]

	if year == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	var params *dynamodb.QueryInput
	if week == "" {
		params = &dynamodb.QueryInput{
			TableName: aws.String("kakor"),
			KeyConditionExpression: aws.String("#yr = :yyyy"),
			ExpressionAttributeNames: map[string]*string{
				"#yr": aws.String("Year"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":yyyy": {
					N: aws.String(year),
				},
			},
		}
	} else {
		params = &dynamodb.QueryInput{
			TableName: aws.String("kakor"),
			KeyConditionExpression: aws.String("#yr = :yyyy and #wk = :week"),
			ExpressionAttributeNames: map[string]*string{
				"#yr": aws.String("Year"),
				"#wk": aws.String("Week"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":yyyy": {
					N: aws.String(year),
				},
				":week": {
					N: aws.String(week),
				},
			},
		}
	}

	result, err := svc.Query(params)
	if err != nil {
		panic(fmt.Sprintf("Failed to build: %v", err))
	}

	var movies []kakItem
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &movies)

	var results []kakResponse2
	for _, m := range movies {
		scores := getScores(m.Year, m.Week)
		results = append(results, kakResponse2{m.Name, m.Year, m.Week, scores})
	}

	s, _ := json.Marshal(results)
	return events.APIGatewayProxyResponse{Body: string(s), StatusCode: 200}, nil


}


func main() {
	lambda.Start(requestHandler)
}
