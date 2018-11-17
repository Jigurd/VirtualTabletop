package diceparse

import (
    "errors"
    "fmt"
    "github.com/Knetic/govaluate"
    "github.com/golang-collections/collections/stack"
    "math/rand"
    "regexp"
    "strconv"
    "strings"
    "time"
)

//Parse searches a string for dicerolls and resolves them. It's very loosely based on ideas from parse trees.
func Parse (input *string) error{
    //go through string character by character, push to stack. If you find "]", pop and append to sting until you find  "[", then resolve diceroll
    strStack := stack.New()

    //check for mismatched brackets
    if strings.Count(*input, "[") != strings.Count(*input, "]"){
        return errors.New("Error: Mismatched brackets")
    }

    //loop until either the string is empty or we've removed all brackets through matching
    for *input != "" {
        if (*input)[:1] != "]" { //if the next character of the string is not a closing bracket
            strStack.Push(TrimChar(input)) //push it to the stack

        } else { //when you find a ]
            TrimChar(input) //get rid of the ]
            var expr string
            for strStack.Len() > 0 && strStack.Peek() != "["{ //go back through the stack until you find a [
                expr = strStack.Pop().(string) + expr //add all characters you find to a new string, in the correct order
            }
            strStack.Pop() // get rid of the [

            roll, err := ParseRoll(expr) //roll the dice specified in the expression
            if err != nil {
                return err
            }

            rollString := strconv.Itoa(roll) //convert that int to a string

            //fmt.Println(rollString) DEBUG

            for len(rollString) > 0 { //push the result back to the stack. This enables nested queries
                strStack.Push(TrimChar(&rollString))
                //fmt.Println(rollString, " ", strStack.Peek()) DEBUG
            }
            //fmt.Println(*input) DEBUG
            if !strings.Contains(*input, "[") && !strings.Contains(*input, "]"){
                break
            }
        }
    }
    result :=""
    for strStack.Len() > 0{ //while stack is not empty
        result= strStack.Pop().(string) + result //add characters to new string in correct order
    }
    fmt.Println(result) //print said string

    return nil
}

//parseRoll parses a string and rolls dice. S'about it. It only supports one "d" per roll at the moment.
func ParseRoll(input string)(result int, err error){
    var left, right interface{}
    var err1, err2 error

    stringIsValid, _ := regexp.MatchString(`([1-9][0-9]*[/+/-/*//])*([1-9][0-9]*)?d?[1-9][0-9]*([/+/-/*//][0-9]+)?`, input)
    if !stringIsValid{
        return result, errors.New("Fatal error - Invalid expression: " +input)
    }


    //small insert to deal with pure math expressions that have no math
    if (!strings.Contains(input, "d")){
        calc, calcErr := evaluate(input)
        if calcErr != nil{
            return 0, err
        }
        return int(calc.(float64)), nil
    }

    //Locate dice in string, find the type of dice being rolled
    sidesStr := regexp.MustCompile("d[0-9]+").FindString(strings.ToLower(input))
    sidesStr = strings.TrimPrefix(sidesStr, "d")                              //strip d as it is unnecessary
     if !isNumeric(sidesStr){                                                  //double-check that remainder is numeric
         fmt.Println("Error parsing diceroll: not numeric")
    }
    sides, _ := strconv.Atoi(sidesStr)                                         //make sides an int so it can be parsed
                                            //error is unecessary here as we already made sure the roll is numeric and such

    //Split input on the dice, to get left and right expressions
    splitInput := regexp.MustCompile("d[0-9]+").Split(strings.ToLower(input), 2)



    // LEFT SIDE
    //if the left part is empty, set it to the default of 1

    //fmt.Println(splitInput[0])

    if splitInput[0] == ""{
        splitInput[0]="1"
    }


    // RIGHT SIDE
    //if the right part is empty, set it to zero to avoid parse errors
    if splitInput[1] == ""{
        splitInput[1] = "0"
    }

    splitInput[1] = "0"+splitInput[1] //add 0 to front of expression, to prevent errors when expression starts with +/-


    //Evaluate expressions
    left, err1 = evaluate(splitInput[0])
    if err1 != nil {
        fmt.Println("Error: Problem with left expression!")
        return 0, err1
    }
    right, err2 = evaluate(splitInput[1])
    if err2 != nil {
        fmt.Println("Error: Problem with right expression!")
        return 0, err2
    }


    //this generates the actual pseudorandom numbers, it generates (left) random numbers in the range [1,sides]
    diceResults, rollError := RollDice(left, sides)

    if rollError != nil{
        return 0, rollError
    }

    result = sumSlice(diceResults)+int(right.(float64))


    fmt.Printf("%v+%v= ", diceResults, right) //DEBUG

    return result, nil
}



//evaluate takes a string and calculates the
func evaluate(s string) (exprReturn interface{}, err error) {
    expr, err := govaluate.NewEvaluableExpression(s)
    if err != nil{
        return nil, err
    }

    exprReturn, err1 :=  expr.Evaluate(nil)

    if err1 != nil{
        return nil, err1
    }

    return exprReturn, nil
}

//RollDice generates "num" random numbers between 1 and "sides", and returns these in a slice
func RollDice (num interface{}, sides int) (result []int, err error){
    if sides <1 {
        return nil, errors.New("Roll Error: Sides can not be <1")
    }

    rand.Seed(time.Now().UnixNano())
    var diceSlice []int

    for i:=0; i<int(num.(float64)); i++{ //convert num from interface to int
        diceSlice = append(diceSlice, 1+rand.Intn(sides))
    }

    return diceSlice, nil
}