package main

import (
	"github.com/urfave/cli"
	"os"
	"log"
	"fmt"
	"galen.demo.go/crawler/epubee"
)

func main() {
	app := cli.NewApp()
	app.Name = "Crawler"
//	app.Usage = "make an explosive entrance"
	app.Action = func(c *cli.Context) error {
		fmt.Println("hello")
		return nil
	}
	/*app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "book, b",
			Value:"",
			Usage: "bookName",
		},
		cli.BoolFlag{
			Name: "all, a",
			Usage: "all",
		},
		cli.StringFlag{
			Name: "outFile, of",
			Value: "out.txt",
			Usage: "outFile",
		},
		cli.BoolFlag{
			Name: "download, d",
			Usage: "download",
		},
	}*/

	app.Commands = []cli.Command{
		{
			Name:    "epubee",
			Aliases: []string{"e"},
			Usage:   "search epubee book list",
			Flags : []cli.Flag {
				cli.StringFlag{
					Name: "book, b",
					Value:"",
					Usage: "bookName",
				},
				cli.BoolFlag{
					Name: "all, a",
					Usage: "all",
				},
				cli.StringFlag{
					Name: "outFile, of",
					Value: "out.txt",
					Usage: "outFile",
				},
				cli.BoolFlag{
					Name: "download, d",
					Usage: "download",
				},
			},
			Subcommands:[]cli.Command{
				{
					Name:    "search",
					Aliases: []string{"s"},
					Usage:   "search book list",
					Action:  func(c *cli.Context) error {
						fmt.Println("begin")
						book := c.GlobalString("book")
						all := c.GlobalBool("all")
						outFile := c.GlobalString("outFile")
						download := c.GlobalBool("download")
						s := c.Args().First()
						if len(s) > 0 {
							book = s
						}
						fmt.Println(book)
						epubee.Retrieve(book,!all,outFile,download)
						return nil
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
