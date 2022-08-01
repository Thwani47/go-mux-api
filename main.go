package main

import (
	"github.com/joho/godotenv"
)

func main() {
	app := App{}
	godotenv.Load(".env")
	app.InitApp(
		GoDotEnvVariable("DEV_HOST"),
		GoDotEnvVariable("DB_USERNAME"),
		GoDotEnvVariable("DB_PASSWORD"),
		GoDotEnvVariable("DB_NAME"))
	app.Run("8001")
}
