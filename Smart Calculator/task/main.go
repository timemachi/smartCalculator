package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var (
	InvalidExprError       = errors.New("invalid expression")
	InvalidAssignmentError = errors.New("invalid assignment")
	InvalidIdentifierError = errors.New("invalid identifier")
	UnknownVariable        = errors.New("unknown variable")
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	scope := make(map[string]int)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch {
		case len(input) == 0:
			continue
		case input == "/help":
			fmt.Println("The program adds and subtracts numbers")
			continue
		case input == "/exit":
			fmt.Println("Bye!")
			return
		case strings.HasPrefix(input, "/"):
			fmt.Println("Unknown command")
			continue
		}

		if isVariableAssigment(input) {
			err := handleVariableAssignment(input, scope)
			printErrorIfKnown(err)
			continue
		}

		tokens, err := parseAndResolveTokens(input, scope)
		if err != nil {
			printErrorIfKnown(err)
		} else {
			ans, err := calculate(tokens)
			if err != nil {
				printErrorIfKnown(err)
			} else {
				fmt.Println(ans)
			}

		}
	}
}

func printErrorIfKnown(e error) {
	switch e {
	case nil:
		return
	case InvalidAssignmentError:
		fmt.Println("Invalid assignment")
	case InvalidIdentifierError:
		fmt.Println("Invalid identifier")
	case UnknownVariable:
		fmt.Println("Unknown variable")
	case InvalidExprError:
		fmt.Println("Invalid expression")
	default:
		log.Fatal(e)
	}
}

// s must contain at least one = sign here
func handleVariableAssignment(s string, scope map[string]int) error {
	split := strings.SplitN(s, "=", 2)
	lhs := strings.TrimSpace(split[0])
	rhs := strings.TrimSpace(split[1])
	if !isValidVariableName(lhs) {
		return InvalidIdentifierError
	}

	num, err := strconv.Atoi(rhs)
	if err == nil {
		scope[lhs] = num
	} else {
		if !isValidVariableName(rhs) {
			return InvalidAssignmentError
		}
		v, err := tryResolve(rhs, scope)
		if err != nil {
			return err
		}
		scope[lhs] = v
	}
	return nil
}

func tryResolve(name string, scope map[string]int) (int, error) {
	if v, ok := scope[name]; !ok {
		return 0, UnknownVariable
	} else {
		return v, nil
	}
}

func isValidVariableName(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

func isVariableAssigment(s string) bool {
	return strings.Contains(s, "=")
}

func calculate(tokens []string) (int, error) {
	myStack := NewStack()
	for _, str := range tokens {
		_, err := strconv.Atoi(str)
		if err == nil {
			myStack.Push(str)
			continue
		}
		val2, err2 := strconv.Atoi(myStack.Peek())
		if err2 != nil {
			return 0, InvalidExprError
		}
		errPop2 := myStack.Pop()
		if errPop2 != nil {
			return 0, errPop2
		}
		val1, err1 := strconv.Atoi(myStack.Peek())
		if err1 != nil {
			return 0, InvalidExprError
		}
		errPop1 := myStack.Pop()
		if errPop1 != nil {
			return 0, errPop1
		}
		result, resultErr := applyOperation(val1, val2, str)
		if resultErr != nil {
			return 0, resultErr
		}
		myStack.Push(strconv.Itoa(result))
	}
	if myStack.Size() == 1 {
		ans := myStack.Peek()
		res, err := strconv.Atoi(ans)
		if err == nil {
			return res, nil
		} else {
			return 0, InvalidExprError
		}
	}
	return 0, InvalidExprError
}

func applyOperation(a, b int, op string) (int, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		return a / b, nil
	case "^":
		return int(math.Pow(float64(a), float64(b))), nil
	default:
		return 0, fmt.Errorf("unknown command: %s", op)
	}
}

// return Postfix expression array
func parseAndResolveTokens(expr string, scope map[string]int) ([]string, error) {
	s := strings.Split(expr, "")
	//resolved is expression without whitespace and variable
	var resolved = ""
	currentVariable := ""
	for _, str := range s {
		//white space
		if str == " " {
			if currentVariable == "" {
				continue
			}
			v, err := tryResolve(currentVariable, scope)
			if err != nil {
				return nil, err
			}
			resolved += strconv.Itoa(v)
			currentVariable = ""
			continue
		}
		//variable name
		if isValidVariableName(str) {
			currentVariable += str
			continue
		}
		//case number, operator, (): check variable; then add curr str
		if currentVariable != "" {
			v, err := tryResolve(currentVariable, scope)
			if err != nil {
				return nil, err
			}
			resolved += strconv.Itoa(v)
			currentVariable = ""

		}
		resolved += str
	}
	if currentVariable != "" {
		v, err := tryResolve(currentVariable, scope)
		if err != nil {
			return nil, err
		}
		resolved += strconv.Itoa(v)
		currentVariable = ""
	}

	//resolved is expression without whitespace and variable, but only digit, operator and parentheses
	res, err := exprToInfixArray(resolved)
	if err != nil {
		return nil, err
	}
	postfixRes, err1 := infixToPostfix(res)
	if err1 != nil {
		return nil, err
	}
	return postfixRes, nil
}

// expr is expression without whitespace and variable, but only digit, operator and parentheses
func exprToInfixArray(expr string) ([]string, error) {
	res := make([]string, 0)
	prevWasNum := false
	prevAdd := 0
	prevSub := 0
	currDigit := ""
	for _, char := range expr {
		if unicode.IsDigit(char) {
			//if prev is not digit, it can be operator or (), check if it's add or sub
			if !prevWasNum {
				if currDigit != "" || (prevAdd != 0 && prevSub != 0) {
					return nil, InvalidExprError
				}
				if prevAdd != 0 {
					res = append(res, "+")
					prevAdd = 0
				}
				if prevSub != 0 {
					if prevSub%2 == 0 {
						res = append(res, "+")
					} else {
						res = append(res, "-")
					}
					prevSub = 0
				}
				currDigit += string(char)
				prevWasNum = true
			} else {
				//if prev is digit: continue to add curr char on the currDigit. But add and sub has to be 0
				currDigit += string(char)
				if prevAdd != 0 && prevSub != 0 {
					return nil, InvalidExprError
				}
			}
		} else { //current char is not a digit, it can be an operator or ()
			//if currDigit is not added to the array, check if it's integer, and then add and reset it
			if currDigit != "" {
				_, err := strconv.Atoi(currDigit)
				if err != nil {
					return nil, InvalidExprError
				}
				res = append(res, currDigit)
				currDigit = ""
			}
			//if char is +, number of sub can't > 0
			if char == '+' {
				if prevSub != 0 {
					return nil, InvalidExprError
				}
				prevAdd++
				if prevWasNum {
					prevWasNum = false
				}
				continue
			}
			//if char is -, number of sub can't > 0
			if char == '-' {
				if prevAdd != 0 {
					return nil, InvalidExprError
				}
				prevSub++
				if prevWasNum {
					prevWasNum = false
				}
				continue
			}
			if char == '*' {
				//prev char has to be a number or ')'
				if !prevWasNum {
					return nil, InvalidExprError
				}
				res = append(res, "*")
				prevWasNum = false
				continue
			}
			if char == '/' {
				if !prevWasNum {
					return nil, InvalidExprError
				}
				res = append(res, "/")
				prevWasNum = false
				continue
			}
			if char == '^' {
				if !prevWasNum {
					return nil, InvalidExprError
				}
				res = append(res, "^")
				prevWasNum = false
				continue
			}
			if char == '(' {
				// char before '(' can't be digit !
				if prevWasNum {
					return nil, InvalidExprError
				}
				if prevAdd != 0 {
					res = append(res, "+")
					prevAdd = 0
				}
				if prevSub != 0 {
					if prevSub%2 == 0 {
						res = append(res, "+")
					} else {
						res = append(res, "-")
					}
					prevSub = 0
				}
				res = append(res, "(")
				continue
			}
			if char == ')' {
				// char before '(' has to be digit !
				if !prevWasNum {
					return nil, InvalidExprError
				}
				res = append(res, ")")
				continue
			}
		}
	}
	if currDigit != "" {
		_, err := strconv.Atoi(currDigit)
		if err != nil {
			return nil, InvalidExprError
		}
		res = append(res, currDigit)
	}
	return res, nil
}

// https://www.geeksforgeeks.org/convert-infix-expression-to-postfix-expression/
func infixToPostfix(expr []string) ([]string, error) {
	result := make([]string, 0, len(expr))
	myStack := NewStack()
	for _, str := range expr {
		// str is digit
		if _, err := strconv.Atoi(str); err == nil {
			result = append(result, str)
			continue
		}
		if str == "(" {
			//push '(' in stack
			myStack.Push(str)
			continue
		}
		if str == ")" {
			for myStack.Size() > 0 && myStack.Peek() != "(" {
				result = append(result, myStack.Peek())
				err := myStack.Pop()
				if err != nil {
					return nil, err
				}
			}
			err := myStack.Pop()
			if err != nil {
				return nil, err
			}
			continue
		}
		// an operator is encountered
		for myStack.Size() > 0 && checkPriority(str) <= checkPriority(myStack.Peek()) {
			result = append(result, myStack.Peek())
			err := myStack.Pop()
			if err != nil {
				return nil, err
			}

		}
		myStack.Push(str)
	}
	for myStack.Size() > 0 {
		if myStack.Peek() == "(" {
			return nil, InvalidExprError
		}
		result = append(result, myStack.Peek())
		err := myStack.Pop()
		if err != nil {
			return nil, err
		}

	}
	return result, nil
}

type Stack struct {
	s []string
}

func NewStack() *Stack {
	return &Stack{make([]string, 0)}
}

func (s *Stack) Push(v string) {
	s.s = append(s.s, v)
}

func (s *Stack) Pop() error {
	l := len(s.s)
	if l <= 0 {
		return InvalidExprError
	}
	s.s = s.s[:l-1]
	return nil
}

func (s *Stack) Peek() string {
	l := len(s.s)
	return s.s[l-1]
}

func (s *Stack) Size() int {
	l := len(s.s)
	return l
}

func checkPriority(expr string) int {
	switch expr {
	case "+":
		return 1
	case "-":
		return 1
	case "*":
		return 2
	case "/":
		return 2
	case "^":
		return 3
	default:
		return -1
	}
}
