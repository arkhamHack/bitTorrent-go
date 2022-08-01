package main

import (
	"log"
	"os"

	"github.com/arkhamHack/bitTorrent-go/torrentCli"
)

func main() {
	input := os.Args[1]
	output := os.Args[2]

	tf, err := torrentCli.Open(input)
	if err != nil {
		log.Fatal(err)
	}
	err = tf.DownloadToFile(output)

	if err != nil {
		log.Fatal(err)
	}

}
