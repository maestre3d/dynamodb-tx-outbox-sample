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

data "aws_iam_policy_document" "assume_lambda" {
  version = "2012-10-17"
  statement {
    effect = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
  }
}

data "aws_iam_policy_document" "log_trailing_daemon" {
  version = "2012-10-17"
  statement {
    sid = "InvokeLambda"
    effect = "Allow"
    actions = ["lambda:InvokeFunction"]
    resources = ["arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:logTrailingDaemon"]
  }
  statement {
    sid = "WriteLogs"
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}"]
  }
  statement {
    sid = "WatchDynamoDbStream"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeStream",
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
      "dynamodb:ListStreams"
    ]
    resources = ["arn:aws:dynamodb:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:table/${aws_dynamodb_table.outbox.name}/stream/*"]
  }
  statement {
    sid = "PublishEvent"
    effect = "Allow"
    actions = ["sns:Publish"]
    resources = ["arn:aws:sns:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
  }
}

resource "aws_iam_role" "log_trailing_daemon" {
  name = "LogTrailingDaemon"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda.json
}
resource "aws_iam_role_policy" "log_trailing_daemon" {
  name = "LogTrailingDaemonOperations"
  role   = aws_iam_role.log_trailing_daemon.id
  policy = data.aws_iam_policy_document.log_trailing_daemon.json
}

resource "aws_sqs_queue" "log_trailing_dlq" {
  name = "logTrailingDaemonDLQ"
}

resource "aws_sqs_queue_policy" "log_trailing_dlq" {
  queue_url = aws_sqs_queue.log_trailing_dlq.url
  policy = jsonencode({
    "Version": "2008-10-17",
    "Id": "__default_policy_ID",
    "Statement": [
      {
        "Sid": "__owner_statement",
        "Effect": "Allow",
        "Principal": {
          "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action": "SQS:*",
        "Resource": aws_sqs_queue.log_trailing_dlq.arn
      },
      {
        "Sid": "EnableLambdaPush",
        "Effect": "Allow",
        "Principal": {
          "AWS": aws_iam_role.log_trailing_daemon.arn
        },
        "Action": [
          "sqs:SendMessage"
        ],
        "Resource": aws_sqs_queue.log_trailing_dlq.arn,
      }
    ]
  })
  depends_on = [aws_sqs_queue.log_trailing_dlq]
}

resource "aws_s3_bucket" "default" {
  bucket = "log-trailing-daemons"
}

resource "aws_lambda_function" "log_trailing_daemon" {
  function_name = "logTrailingDaemon"
  role          = aws_iam_role.log_trailing_daemon.arn
  handler = "main"
  runtime = "go1.x"
  s3_bucket = aws_s3_bucket.default.bucket
  // REQUIRED TO UPLOAD AN ARTIFACT, OTHERWISE THE LAMBDA FUNCTION DEPLOYMENT WILL FAIL
  s3_key = "log_trailing_daemon.zip"
  dead_letter_config {
    target_arn = aws_sqs_queue.log_trailing_dlq.arn
  }
  depends_on = [aws_iam_role.log_trailing_daemon, aws_sqs_queue.log_trailing_dlq, aws_s3_bucket.default]
}

resource "aws_lambda_event_source_mapping" "outbox_stream" {
  function_name = aws_lambda_function.log_trailing_daemon.arn
  event_source_arn = aws_dynamodb_table.outbox.stream_arn
  starting_position = "LATEST"
  maximum_retry_attempts = 30
  destination_config {
    on_failure {
      destination_arn = aws_sqs_queue.log_trailing_dlq.arn
    }
  }

  depends_on = [aws_lambda_function.log_trailing_daemon]
}

resource "aws_sns_topic" "student_registered" {
  name = "ncorp-workspaces-iam-1-student-registered"
  sqs_failure_feedback_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/SNSFailureFeedback"
  sqs_success_feedback_role_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/SNSSuccessFeedback"
  sqs_success_feedback_sample_rate = 30
  policy = jsonencode({
    "Version": "2008-10-17",
    "Id": "__default_policy_ID",
    "Statement": [
      {
        "Sid": "__default_statement_ID",
        "Effect": "Allow",
        "Principal": {
          "AWS": "*"
        },
        "Action": [
          "SNS:Publish",
          "SNS:RemovePermission",
          "SNS:SetTopicAttributes",
          "SNS:DeleteTopic",
          "SNS:ListSubscriptionsByTopic",
          "SNS:GetTopicAttributes",
          "SNS:AddPermission",
          "SNS:Subscribe"
        ],
        "Resource": "arn:aws:sns:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:ncorp-workspaces-iam-1-student-registered",
        "Condition": {
          "StringEquals": {
            "AWS:SourceOwner": data.aws_caller_identity.current.account_id
          }
        }
      },
      {
        "Sid": "__console_pub_0",
        "Effect": "Allow",
        "Principal": {
          "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action": "SNS:Publish",
        "Resource": "arn:aws:sns:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:ncorp-workspaces-iam-1-student-registered"
      },
      {
        "Sid": "__console_sub_0",
        "Effect": "Allow",
        "Principal": {
          "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action": "SNS:Subscribe",
        "Resource": "arn:aws:sns:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:ncorp-workspaces-iam-1-student-registered"
      }
    ]
  })
}

resource "aws_sqs_queue" "log_on_student_registered" {
  name = "log-on-student_registered"
}

resource "aws_sqs_queue_policy" "log_on_student_registered" {
  queue_url = aws_sqs_queue.log_on_student_registered.id
  policy = jsonencode({
    "Version": "2008-10-17",
    "Id": "__default_policy_ID",
    "Statement": [
      {
        "Sid": "__owner_statement",
        "Effect": "Allow",
        "Principal": {
          "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action": "SQS:*",
        "Resource": aws_sqs_queue.log_on_student_registered.arn
      },
      {
        "Sid": "__sender_statement",
        "Effect": "Allow",
        "Principal": {
          "AWS": "*"
        },
        "Action": "SQS:SendMessage",
        "Resource": aws_sqs_queue.log_on_student_registered.arn,
        "Condition": {
          "ArnLike": {
            "aws:SourceArn": aws_sns_topic.student_registered.arn
          }
        }
      },
      {
        "Sid": "__receiver_statement",
        "Effect": "Allow",
        "Principal": {
          "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action": [
          "SQS:ChangeMessageVisibility",
          "SQS:DeleteMessage",
          "SQS:ReceiveMessage"
        ],
        "Resource": "arn:aws:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-on-student_registered"
      }
    ]
  })
  depends_on = [aws_sns_topic.student_registered, aws_sqs_queue.log_on_student_registered]
}

resource "aws_sns_topic_subscription" "log_on_student_registered" {
  endpoint  = aws_sqs_queue.log_on_student_registered.arn
  protocol  = "sqs"
  topic_arn = aws_sns_topic.student_registered.arn
  depends_on = [aws_sqs_queue.log_on_student_registered, aws_sns_topic.student_registered]
}
