package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"

	"github.com/go-kit/kit/log"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "client":
			var args []string
			if len(os.Args) > 2 {
				args = os.Args[2:]
			}
			runClient(args)
		case "server":
			var args []string
			if len(os.Args) > 2 {
				args = os.Args[2:]
			}
			runServer(args)
		}
	}
}

func runServer(args []string) {
	flagset := flag.NewFlagSet("server", flag.ExitOnError)
	var (
		flAddr = flagset.String("addr", ":8080", "server address")
	)
	if err := flagset.Parse(args); err != nil {
		fatal(err)
	}

	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		io.Copy(os.Stdout, r.Body)
	})

	if err := http.ListenAndServe(*flAddr, nil); err != nil {
		fatal(err)
	}
}

type config struct {
	URL string `json:"url"`
}

func runClient(args []string) {
	flagset := flag.NewFlagSet("client", flag.ExitOnError)
	var (
		flConfig = flagset.String("config", "/etc/micromdm/ulog/server.json", "path to config file")
	)
	if err := flagset.Parse(args); err != nil {
		fatal(err)
	}

	data, err := ioutil.ReadFile(*flConfig)
	if err != nil {
		fatal(err)
	}
	var conf config
	if err := json.Unmarshal(data, &conf); err != nil {
		fatal(err)
	}

	r, w := io.Pipe()
	logger := log.NewJSONLogger(w)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		req, err := http.NewRequest("POST", conf.URL, r)
		if err != nil {
			fatal(err)
		}
		reqctx := req.WithContext(ctx)
		if resp, err := http.DefaultClient.Do(reqctx); err != nil {
			fatal(err)
		} else {
			io.Copy(os.Stdout, resp.Body)
		}
	}()
	if err := startLogReader(ctx, logger); err != nil {
		fatal(err)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig
}

func startLogReader(ctx context.Context, logger log.Logger) error {
	cmd := exec.CommandContext(ctx, "/usr/bin/log", "stream", "--level=debug")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go copyLogs(ctx, io.MultiReader(stdout, stderr), logger)
	return cmd.Start()
}

func copyLogs(ctx context.Context, r io.Reader, logger log.Logger) {
	rdr := bufio.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			logger.Log("err", ctx.Err())
		default:
			line, _, err := rdr.ReadLine()
			if err != nil {
				logger.Log("err", err)
				goto done
			}
			logger.Log("msg", string(line))
		}
	}
done:
	return
}

func fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}
