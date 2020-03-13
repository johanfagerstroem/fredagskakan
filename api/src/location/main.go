package main

import (
	"fmt"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type scoreItem struct {
	YearWeekLocation string
	Year int
	Week int
	Location string
	Score int
}

type response struct {
	Locations []string
}



func requestHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	sess := session.Must(session.NewSession())
	svc  := dynamodb.New(sess)


	params := &dynamodb.ScanInput{
		TableName: aws.String("votes"),
	}

	result, err := svc.Scan(params)
	if err != nil {
		panic(fmt.Sprintf("Failed to scan: %v", err))
	}

	var scores []scoreItem
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &scores)

	var locations map[string]int
	locations = make(map[string]int)
	for _, s := range scores {
		locations[s.Location] = locations[s.Location] + 1
	}

	res := response{nil}
	for k := range locations {
		res.Locations = append(res.Locations, k)
	}

	s, _ := json.Marshal(res)

	return events.APIGatewayProxyResponse{Body: string(s), StatusCode: 200}, nil

}


func main() {
	lambda.Start(requestHandler)
}
