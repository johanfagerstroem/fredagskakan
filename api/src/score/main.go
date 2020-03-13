package main

import (
	"fmt"
	"time"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type scoreItem struct {
	YearWeekLocation string
	Year int
	Week int
	Location string
	ScoreSum int
	VoteCount int
}

type scoreResponse struct {
	Location string
	Score int
}



func requestHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	sess := session.Must(session.NewSession())
	svc  := dynamodb.New(sess)

	year, week := time.Now().ISOWeek()

	location := request.PathParameters["location"]

	filtYear := expression.Name("Year").Equal(expression.Value(year))
	filtWeek := expression.Name("Week").Equal(expression.Value(week))
	filtLocation := expression.Name("Location").Equal(expression.Value(location))

	filter := filtYear.And(filtWeek)
	if location != "" {
		filter = filtYear.And(filtWeek.And(filtLocation))
	}

	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to build: %v", err))
	}

	params := &dynamodb.ScanInput{
		TableName: aws.String("votes"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	result, err := svc.Scan(params)
	if err != nil {
		panic(fmt.Sprintf("Failed to scan: %v", err))
	}

	var scores []scoreItem
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &scores)

	if location != "" {
		res := scoreResponse{scores[0].Location, scores[0].ScoreSum/scores[0].VoteCount}
		s, _ := json.Marshal(res)
		return events.APIGatewayProxyResponse{Body: string(s), StatusCode: 400}, nil
	} else {
		totalScoreSum := 0
		totalVoteCount := 0
		avg := 0
		for _, s := range scores {
			totalScoreSum = totalScoreSum + s.ScoreSum
			totalVoteCount = totalVoteCount + s.VoteCount
		}
		if totalVoteCount > 0 {
			avg = totalScoreSum / totalVoteCount
		}
		res := scoreResponse{"all", avg}
		s, _ := json.Marshal(res)
		return events.APIGatewayProxyResponse{Body: string(s), StatusCode: 400}, nil
	}

}


func main() {
	lambda.Start(requestHandler)
}
