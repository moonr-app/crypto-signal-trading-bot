# Crypto-signal-trading-bot

## Rationalè
As part of our research for creating [moonr.app](https://moonr.app), we wanted to understand just how hard it is to try
and time the market when buying cryptocurrencies and how for most people this is not the right strategy. After making a bit of money & losing a lot of money, 
those lessons have definitely been learned.

If you find this project interesting, please head over to [MoonR](https://moonr.app)
and enter your details to get regular updates as we get ready to launch.

## What Exactly is this project?
This bot was inspired by [this](https://github.com/CyberPunkMetalHead/gateio-crypto-trading-bot-binance-announcements-new-coins) python project. However, this is a pure Go implementation written from the ground up and with some extra
features.

Once deployed, this bot scrapes Coinbase's API and Binance's coin announcement blog on a specified interval to look for newly listed coins.
Once it finds a new coin, the bot will make a purchase on [gate.io](https://www.gate.io/ref/7618463) (if you don't have an account please sign up 
using this link to support this project. You will also get a discount on fees). The goal is to make the purchase at as close to the announcement time 
as possible. The bot will then check the price of the coin on gate.io at a specified interval, and sell the coins if it goes above a specified threshold.

There is currently no implementation of a stop loss, so you'll need to step in and manually sell the coins if you do not buy at the right time or it never
reaches your threshold.

# Getting started
To get Started you'll need:
- [A gate.io account](https://www.gate.io/ref/7618463)
- [An AWS account](https://console.aws.com). If you do everything right, this whole project can be run on the free tier.
- Some knowledge of AWS/Go. However, we have tried to explain everything in as much detail as possible.

## Project Structure
The main binaries in this project are all housed in `cmd`. These are:
- **coinbasescratch**: This is a test project for testing the integration with coinbase. Its meant to be a playground for you to get comfortable with the coinbase integration in isolation.
- **dbscratch**: Same as above but for testing integration with Dynamo.
- **gateioscratch**: Same as above but for gate.io. Note if it doesn't run in test mode, it really will buy and sell coins.
- **tradebot**: This is the real app binary for the trade bot.


# Deploying

## Two Warnings Before You Start:
- The amount to buy is specified in the env var `USDT_TO_SPEND`. If you don't have enough money in your gate.io account, the bot will fail to buy.
- If you have a clean DB and the last post on binance happens to be a new listing, the bot will buy the coins. We recommend deploying it in
test mode the first time you run it to protect against that. 

## Database
The bot is currently setup to work with Dynamo DB. It uses Dynamo to track the status of the coins it has and has not purchased.
To setup a Dynamo DB and learn more about it, you can follow instructions [here](https://pages.awscloud.com/AWS-Learning-Path-Getting-Started-with-Amazon-DynamoDB_2020_LP_0004-DAT.html).


We recommend creating the DB in the region you intend to deploy the bot in. We did some benchmark tests of where we found the best place
to run the bot, you can see them [here](./benchmarks.md). However, the TLDR is Seoul or Singapore is a good bet.


Once you have created a dynamoDB, create a table called `coin_history`. If for whatever reason you don't want to call it `coin_history`, you'll need to edit
`internal/persistence/dynamodb.go`.


## gate.io
After that, you need to get API Keys from gate.io. You can find instruction on how to do that [here](https://support.gate.io/hc/en-us/articles/900000114363-What-are-APIKey-and-APIV4keys-for-).

## Telegram
This bot has the ability to write to telegram each time it buys and sells. To do this you need to simply update the telegram config
in `internal/notifier/telegram.go`. You can find more about writing to telegram [here](https://core.telegram.org/bots/api).

## Rest of Env Vars

Next, create an env file based on `.env.example` and fill in the values. Comments below for what each env does
```
SELL_THRESHOLD_PERCENTAGE= #what percentage increase to sell at. 20 would sell at a 20% increase.
GATE_API_KEY= #obvious
GATE_API_SECRET= #obvious
DISABLE_TELEGRAM=false #if true, the bot won't write to the telegram channel when it buys and sells
ENABLE_TEST_MODE=false # if true, the bot won't actually buy or sell. Will still write to the db as if it did
DYNAMO_ID= #you get this from AWS
DYNAMO_SECRET= #you get this from AWS
DYNAMO_REGION=eu-west-2 #must be correct for where you created your dynamo db. I reccomend putting this in the same region you intend to deploy bot.
BUY_INTERVAL_SECONDS=1 #interval to check whether to buy (in seconds)
SEll_INTERVAL_SECONDS=1 #interval to check whether to sell (in seconds)
BOT_OWNER= #your name
USDT_TO_SPEND= #amount you want to spend each run per coin.
TICKER_CACHE_INTERVAL_SECONDS=#of seconds to cache prices.
```

## EC2
Create an EC2 instance in the AWS console. We don't need anything beefy so whatever is within the free tier is fine.
We recommend creating it in the same region as your dynamo DB.

We used `Amazon Linux 2 AMI`. Once your instance is created, ssh into it. 

Once you are sshed, run
```
sudo yum update -y
```
then:
```
sudo amazon-linux-extras install docker
```

then:
```
 sudo service docker start
```

then:
```
 sudo usermod -a -G docker ec2-user
```
then:
````
logout 
````
ssh back into your terminal (you have to logout for docker change to take effect).

then:
```
sudo yum install git-all
```

We now have git and docker installed. Add your ssh-keys to your instance so that you have the right to pull from the git repo.

clone the repo:
```
git clone git@github.com:moonr-app/crypto-signal-trading-bot.git
```

copy your env file onto the server. Now run:
```
 docker build . -t crypto-trade-bot:01-31-2022
```
and run it in detached mode:
```
docker run -d -t  --env-file=.env crypto-trade-bot:01-31-2022
```

to "deploy" a new version, kill the docker container and run the last 2 commands again.

## Configuring Cloudwatch Logs:

The following IAM role will need to be added to EC2 instance:

(`EC2 Instance dashboard` -> `Actions` -> `Security` -> `Modify IAM Role`)

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "logs:CreateLogStream",
                "logs:PutLogEvents",
                "logs:CreateLogGroup"
            ],
            "Effect": "Allow",
            "Resource": "*"
        }
    ]
}
```

To run docker image and export logs to cloudwatch run:
```
docker run -it --log-driver=awslogs --log-opt awslogs-region=ap-northeast-1  --log-opt awslogs-group=cryptoBotGroup --log-opt awslogs-create-group=true -d -t --env-file=.env crypto-trade-bot:01-31-2022
```
(replace `ap-northeast-1` with relevant region)

This will create a log group called `cryptoBotGroup` in the same region.


## Warning
This project has the ability to spend real money so please ensure you read through the code and understand how it works before 
committing to running it. We cannot be held responsible for any losses incurred because of its use.
