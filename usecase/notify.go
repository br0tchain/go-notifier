package usecase

import (
	"bufio"
	"fmt"
	"github.com/br0tchain/go-notifier/domain"
	"github.com/br0tchain/go-notifier/lib"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"time"
)

var (
	inputRegexp        = regexp.MustCompile(`^notify --url[ =]([\S]+) (.*)$`)
	helpRegexp         = regexp.MustCompile(`^notify --help`)
	flagIntervalRegexp = regexp.MustCompile(`-i[ =]([0-9]+[a-z]+)`)
	flagMessageRegexp  = regexp.MustCompile(`-m[ =]([\S]+])`)
	flagFileRegexp     = regexp.MustCompile(`-f[ =]([\S]+])`)
)

type notifier struct {
	Client lib.Client
}

type Notify interface {
	ReadInput()
}

func NewNotifier(client lib.Client) Notify {
	return notifier{
		Client: client,
	}
}

func (n notifier) ReadInput() {
	displayHelp()
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			err := n.parseInput(text)
			if err != nil {
				fmt.Printf("An error occurred: %s\n", err.Error())
				fmt.Println("Provide an URL and a message content to send")
			}
		}

		if scanner.Err() != nil {
			// Handle error.
			fmt.Println("Provide an URL and a message content to send")
		}
		return
	}()
}

func (n notifier) parseInput(input string) error {
	submatch := inputRegexp.FindAllStringSubmatch(input, -1)
	//no url specified
	if len(submatch) == 0 {
		help := helpRegexp.FindAllString(input, -1)
		//requesting help
		if len(help) > 0 {
			displayHelp()
		} else { //invalid input
			displayInvalidInput()
		}
	}
	dataToHandle := submatch[0]
	if len(dataToHandle) < 3 {
		displayInvalidInput()
	}

	params, err := parseFlags(dataToHandle[2])
	if err != nil {
		return err
	}
	request, err := n.Client.Prepare(dataToHandle[1], params.Body)
	if err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(params.Interval)
		defer ticker.Stop()
		for {
			results := n.Client.Notify(request)
			result := n.Client.GetNotificationResult(results)
			if result.IsError {
				fmt.Printf("warning: error occurred on request to \n %s \n with response \n %+v", request.URL.String(), result.ErrorDetails)
			}
			<-ticker.C
		}
	}()
	return nil
}

func parseFlags(flags string) (*domain.Params, error) {
	interval, _ := time.ParseDuration("5s")
	//parsing interval
	intervalParsed := flagIntervalRegexp.FindAllStringSubmatch(flags, -1)
	//overriding default interval
	if len(intervalParsed) > 0 {
		i, err := time.ParseDuration(intervalParsed[0][1])
		if err != nil {
			return nil, err
		}
		interval = i
	}

	//parsing message
	messageParsed := flagMessageRegexp.FindAllStringSubmatch(flags, -1)
	if len(messageParsed) > 0 {
		return &domain.Params{
			Interval: interval,
			Body:     messageParsed[0][1],
		}, nil
	}
	//no message, parsing file
	fileParsed := flagFileRegexp.FindAllStringSubmatch(flags, -1)
	if len(fileParsed) == 0 {
		return nil, fmt.Errorf("no message content has been provided")
	}
	data, err := os.ReadFile(fileParsed[0][1])
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("file %s could not be read", fileParsed[0][1]))
	}
	return &domain.Params{
		Interval: interval,
		Body:     string(data),
	}, nil
}

func displayHelp() {
	fmt.Println(`usage: notify --url=URL [<flags>]
Flags:
	--help 	Show context-sensitive help.
	-i      =5s Notification interval.
    -m 		Specify message to be sent 
    -f 		Retrieve the content of a file to be sent as the message content 
Example call:
	$ notify --url http://localhost:8080/notify --file messages.txt`)
}

func displayInvalidInput() {
	fmt.Println(`Invalid input, please provide a valid input or type notify --help for help`)
}
