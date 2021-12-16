provider "aws" {
  region  = "us-east-1"

  default_tags {
    tags = {
      platform    = "test_outbox"
    }
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.58.0"
    }
  }
}

resource "aws_dynamodb_table" "students" {
  name     = "students"
  hash_key = "student_id"
  range_key = "school_id"
  read_capacity = 5
  write_capacity = 5

  attribute {
    name = "student_id"
    type = "S"
  }

  attribute {
    name = "school_id"
    type = "S"
  }
}

resource "aws_dynamodb_table" "outbox" {
  name     = "outbox"
  hash_key = "transaction_id"
  range_key = "occurred_at"
  read_capacity = 5
  write_capacity = 5

  attribute {
    name = "transaction_id"
    type = "S"
  }

  attribute {
    name = "occurred_at"
    type = "S"
  }

  ttl {
    attribute_name = "time_to_exist"
    enabled = true
  }

  stream_enabled = true
  stream_view_type = "NEW_IMAGE"
}