package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/maharshiaj/f1-strategy/internal/openf1"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: f1 <command> [flags]")
	}

	switch args[0] {
	case "sessions":
		return runSessions(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runSessions(args []string) error {
	client := openf1.NewClient()

	var filter openf1.SessionsFilter

	fs := flag.NewFlagSet("sessions", flag.ContinueOnError)

	fs.IntVar(&filter.Year, "year", 0, "the year of the session to get")
	fs.StringVar(&filter.Type, "type", "", "the type of session to get")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if filter.Year == 0 {
		return errors.New("--year is required")
	}

	ctx := context.Background()

	sessions, err := client.Sessions(ctx, filter)

	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("no sessions found")
		return nil
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartDate.Compare(sessions[j].StartDate) < 0
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	for _, session := range sessions {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", session.SessionKey, session.StartDate.Format("2006-01-02"), session.CountryName, session.Location, session.SessionName)
	}

	return w.Flush()
}
