package usecase

import (
	"bufio"
	"fmt"
	"github.com/br0tchain/go-notifier/domain"
	"github.com/br0tchain/go-notifier/lib"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	defaultDuration        = "5s"
	errNoContent           = "no message content has been provided"
	errInvalidInput        = "invalid input"
	errTemplateNoFileFound = "file %s could not be read"
)

var (
	inputRegexp        = regexp.MustCompile(`^notify --url[ =](\S+) (.*)$`)
	helpRegexp         = regexp.MustCompile(`^notify --help`)
	flagIntervalRegexp = regexp.MustCompile(`-i[ =]([0-9]+[a-z]+)`)
	flagMessageRegexp  = regexp.MustCompile(`-m[ =]"(.*)"`)
	flagFileRegexp     = regexp.MustCompile(`-f[ =](\S+)`)
	flagSilentRegexp   = regexp.MustCompile(`--silent`)
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

//run the scanner to get the input from the user for different notifications
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

//parse the input received to extract the url and other params to send the notification
func (n notifier) parseInput(input string) error {
	submatch := inputRegexp.FindAllStringSubmatch(input, -1)
	//no url specified
	if len(submatch) == 0 {
		help := helpRegexp.FindAllString(input, -1)
		//requesting help
		if len(help) > 0 {
			displayHelp()
			return nil
		} else { //invalid input
			displayInvalidInput()
			return fmt.Errorf(errInvalidInput)
		}
	}
	dataToHandle := submatch[0]
	//check if enough params provided
	if len(dataToHandle) < 3 {
		displayInvalidInput()
		return fmt.Errorf(errInvalidInput)
	}
	//retrieving flags from command
	params, err := n.parseFlags(dataToHandle[2])
	if err != nil {
		return err
	}
	//prepare http request
	request, err := n.Client.Prepare(dataToHandle[1], params.Body)
	if err != nil {
		return err
	}
	//sending notification
	go n.triggerNotification(params, request)
	return nil
}

//trigger notification to send the request every interval
func (n notifier) triggerNotification(params *domain.Params, request *http.Request) {
	ticker := time.NewTicker(params.Interval)
	defer ticker.Stop()
	for {
		//sending notification
		results := n.Client.Notify(request)
		if !params.IsSilent {
			go n.getError(results, request)
		}
		//waiting for next tick to loop
		<-ticker.C
	}
}

//retrieve the error of the request from the results channel
func (n notifier) getError(results <-chan *lib.NotificationResult, request *http.Request) {
	//retrieving notification once received
	result := n.Client.GetNotificationResult(results)
	if result.IsError {
		fmt.Printf("warning: error occurred on request to \n %s \n with response \n %+v\n", request.URL.String(), result.ErrorDetails)
	}
}

//parse the flags receive to extract the body of the notification and the interval requested
func (n notifier) parseFlags(flags string) (*domain.Params, error) {
	interval, _ := time.ParseDuration(defaultDuration)
	isSilent := false
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

	//no message, parsing file
	isSilentParsed := flagSilentRegexp.FindAllStringSubmatch(flags, -1)
	if len(isSilentParsed) > 0 {
		isSilent = true
	}
	//parsing message
	messageParsed := flagMessageRegexp.FindAllStringSubmatch(flags, -1)
	if len(messageParsed) > 0 {
		return &domain.Params{
			Interval: interval,
			Body:     messageParsed[0][1],
			IsSilent: isSilent,
		}, nil
	}
	//no message, parsing file
	fileParsed := flagFileRegexp.FindAllStringSubmatch(flags, -1)
	if len(fileParsed) == 0 {
		return nil, fmt.Errorf(errNoContent)
	}
	data, err := os.ReadFile(fileParsed[0][1])
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(errTemplateNoFileFound, fileParsed[0][1]))
	}
	return &domain.Params{
		Interval: interval,
		Body:     string(data),
		IsSilent: isSilent,
	}, nil
}

func displayHelp() {
	fmt.Println(`usage: notify --url=URL [<flags>]
Flags:
	--help		Show context-sensitive help.
	-i		=5s Notification interval.
	-m		Specify message to be sent 
	-f		Retrieve the content of a file to be sent as the message content 
	--silent	Add this flag to avoid displaying error messages
Example call:
	$notify --url http://localhost:8080/notify -m "content to be sent" --silent`)
}

func displayInvalidInput() {
	fmt.Println(`Invalid input, please provide a valid input or type notify --help for help`)
}
