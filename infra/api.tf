resource "aws_dynamodb_table" "kakor" {
  name         = "kakor"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "Year"
  range_key    = "Week"

  attribute {
    name = "Year"
    type = "N"
  }

  attribute {
      name = "Week"
      type = "N"
  }
}

resource "aws_dynamodb_table" "votes" {
  name         = "votes"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "YearWeekLocation"

  attribute {
    name = "YearWeekLocation"
    type = "S"
  }
}