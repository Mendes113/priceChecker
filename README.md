# Amazon Price Tracker

A simple Amazon price tracker written in Go. This tool allows you to monitor the prices of products on Amazon and track variations over time.

## Features

- Fetches product prices from Amazon
- Tracks price variations and historical data
- Saves data to a CSV file for future analysis
- Sends email notifications for significant price changes

## Prerequisites

Before running the program, make sure you have the following dependencies installed:

- Go (Programming Language)
- PuerkitoBio/goquery (HTML parser)
- SMTP server credentials (for email notifications)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/amazon-price-tracker.git
cd amazon-price-tracker
```
Install dependencies:
bash
Copy code
go get -u github.com/PuerkitoBio/goquery
Usage
Modify the main function in main.go with the URL of the product you want to track.

Run the program:

bash
Copy code
go run main.go
The program will fetch the product's current price, calculate variations, and save the data to a CSV file.

Configuration
Configure email settings in the setupEmail function if you want to receive notifications.

go
Copy code
from := "your-email@gmail.com"
password := "your-email-password"
to := "recipient@example.com"
License
This project is licensed under the MIT License - see the LICENSE file for details.

sql
Copy code

Remember to replace placeholders like `yourusername` and update the email config

