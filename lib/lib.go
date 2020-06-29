package histweet

import (
	"errors"
	"log"
)

func runSingle(args *Args) error {
	log.Println("Called runSingle")

	_, err := NewTwitterClient(args)
	if err != nil {
		return errors.New("Failed to build Twitter client")
	}

	return nil
}

func runDaemon(args *Args) error {
	log.Println("Called runDaemon")
	return nil
}

func Run(args *Args) error {
	if args.Daemon {
		return runDaemon(args)
	} else {
		return runSingle(args)
	}
}
