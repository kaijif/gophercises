package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"time"
)

type Problem struct {
	question string
	answer   string
}

func (p Problem) IsCorrect(usr_ans string) bool {
	return p.answer == usr_ans
}

func readCSV(filePath *string) []Problem {
	f, err := os.Open(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Uh oh, we couldn't find the problems. Apparently, %s", err.Error())
		os.Exit(1)
	}
	defer f.Close()

	// readcsv
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Uh oh, we couldn't read the problems. Apparently, %s", err.Error())
		os.Exit(1)
	}

	// transmute records into Problem array
	problems := make([]Problem, 0, 100)

	for _, record := range records {
		problem := Problem{question: record[0], answer: record[1]}
		problems = append(problems, problem)
	}

	return problems
}

// struct for storing quiz stats

func (p Problem) askAQuestion(cookies *int, i int) bool {
	fmt.Printf("Question %d: ", i)
	fmt.Println(p.question)

	// get user answer
	var usr_ans string
	fmt.Scanln(&usr_ans)
	correct := p.IsCorrect(usr_ans)
	if correct {
		fmt.Println("Correct! You get a cookie!!")
		*cookies += 1
	} else {
		fmt.Println("Wrong! No cookie for you...")
	}
	return correct
}

func runQuiz(problems []Problem, done chan bool, timer_done <-chan time.Time, cookies *int) {

	for i, problem := range problems {
		select {
		case <-timer_done:
			return
		default:
			problem.askAQuestion(cookies, i)
		}
	}

	done <- true
}

func wrapUp(cookies int, timed_out bool) {
	if timed_out {
		fmt.Println("Womp womp... ran out of time")
	}
	fmt.Println("Quiz finished! Here are your stats:")
	fmt.Printf("Cookies: %d\n", cookies)
}

func main() {
	filePath := flag.String("filePath", "./problems.csv", "Path to CSV file containing questions")
	sleepLength := flag.Int("quizLength", 30, "Quiz duration in seconds")

	// load in questions
	problems := readCSV(filePath)

	// make channels, cookies
	quiz_done := make(chan bool)
	var cookies int

	// start the timer
	timer := time.NewTimer(time.Second * time.Duration(*sleepLength))

	// start the quiz
	go runQuiz(problems, quiz_done, timer.C, &cookies)

	select {
	case <-quiz_done:
		wrapUp(cookies, false)
	case <-timer.C:
		wrapUp(cookies, true)
	}

	return
}
