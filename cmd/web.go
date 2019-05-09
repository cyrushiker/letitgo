package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/urfave/cli"
	"gopkg.in/macaron.v1"

	"github.com/cyrushiker/letitgo/pkg/setting"
)

// Web to run a web server
var Web = cli.Command{
	Name:  "web",
	Usage: "Start web server",
	Description: `You will need to set host and ip first before run,
	otherwise it will runing with default setting.`,
	Action: runWeb,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "port, p", Value: "9000", Usage: "Temporary port number to prevent conflict"},
		cli.StringFlag{Name: "config, c", Value: "custom/conf/app.ini", Usage: "Custom configuration file path"},
	},
}

func runWeb(c *cli.Context) error {
	setting.NewContext()
	if c.IsSet("port") {
		setting.AppURL = strings.Replace(setting.AppURL, setting.HTTPPort, c.String("port"), 1)
		setting.HTTPPort = c.String("port")
	}
	listenAddr := fmt.Sprintf("%s:%s", setting.HTTPAddr, setting.HTTPPort)

	// macaron routers
	m := macaron.Classic()
	m.Use(macaron.Renderer())
	// static file
	m.Use(macaron.Static(
		"public",
		macaron.StaticOptions{
			SkipLogging: setting.DisableRouterLog,
		},
	))

	m.Get("/", func(ctx *macaron.Context) {
		ctx.Data["Website"] = "cyrushiker.me"
		ctx.Data["Email"] = "cyrushiker@outlook.com"
		ctx.HTML(200, "index") // 200 is the response code.
	})

	m.Get("/l", longHander)
	m.Get("/s", shortHander)
	log.Printf("Server is runing on: %s\n", setting.AppURL)
	log.Fatal(http.ListenAndServe(listenAddr, m))
	return nil
}

// ServerResult the server result type
type ServerResult struct {
	Code    int                    `json:"code"`
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
}

func longTask() {
	start := time.Now()
	log.Println("I will run for at least 1 min")
	time.Sleep(time.Second * 5)
	ch := make(chan string)
	urls := []string{"https://git.1mdata.com", "http://gopl.io", "https://dc.1mdata.com"}
	for _, url := range urls {
		go fetch(url, ch)
	}
	for range urls {
		fmt.Println(<-ch)
	}
	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
	log.Println("task is done")
}

func longHander(ctx *macaron.Context) string {
	log.Println("work on long task")
	go longTask()
	return "long task is running"
}

func shortTask() {
	log.Println("I will return back immidiately with nothing to do")
}

func shortHander(ctx *macaron.Context) string {
	log.Println("work on short task")
	go shortTask()
	ctx.Header().Set("Content-Type", "application/json")
	sr := ServerResult{Code: 200, Success: true, Data: nil}
	srs, err := ctx.JSONString(sr)
	if err != nil {
		log.Println(err)
	}
	return srs
}

func fetch(url string, ch chan<- string) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	nbytes, err := io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	if err != nil {
		ch <- fmt.Sprintf("whild reading %s: %v", url, err)
		return
	}
	secs := time.Since(start).Seconds()
	ch <- fmt.Sprintf("%.2fs %7d  %s", secs, nbytes, url)
}
