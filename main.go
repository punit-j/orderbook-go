package main

import (
	"fmt"
	"orders-manager/config"
	"orders-manager/server"
	"orders-manager/service"
	"os"
	"strconv"
	"time"

	logger "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"github.com/urfave/negroni"
)

func main() {
	config.Load()

	cliApp := cli.NewApp()
	cliApp.Name = config.AppName()
	cliApp.Version = "1.0.0"
	cliApp.Commands = []*cli.Command{
		{
			Name:  "start",
			Usage: "start server",
			Action: func(c *cli.Context) error {
				return startApp()
			},
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		panic(err)
	}
}
func startApp() (err error) {
	deps, err := server.InitDependencies()

	_, handler := server.InitRouter(deps)

	go func() {
		for {
			// Finding Matches
			matches, err := service.FindMatches()
			if err != nil {
				logger.Error(err)
			} else {
				service.TransferTokens(matches)
			}
			logger.Info("Sleeping for 10 seconds")
			// Sleep for 10 seconds
			time.Sleep(10 * time.Second)
		}
	}()
	// init web server
	server := negroni.Classic()
	server.UseHandler(handler)

	port := config.AppPort() // This can be changed to the service port number via environment variable.
	addr := fmt.Sprintf(":%s", strconv.Itoa(port))

	server.Run(addr)
	return
}
