#!/bin/bash
awslocal sqs create-queue --queue-name weather
awslocal s3 mb s3://weather-raw
