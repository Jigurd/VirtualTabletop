package diceparse

import (
    "errors"
    "fmt"
    "github.com/Knetic/govaluate" //
    "math/rand"
    "regexp"
    "strconv"
    "strings"
    "time"
)

//Parse parses a string and rolls dice. S'about it. It only supports one "d" per roll at the moment.
func Parse(input string)(result int, err error){
    var left, right interface{}
    var err1, err2 error

    stringIsValid, _ := regexp.MatchString(`([0-9]+[/+/-/*//])?[0-9]+d[0-9]+([/+/-/*//][0-9]+)?`, input)
    if !stringIsValid{
        return result, errors.New("Fatal error: Invalid expression.")
    }

    //Locate dice in string, find the type of dice being rolled
    sidesStr := regexp.MustCompile("d[0-9]+").FindString(strings.ToLower(input))
    sidesStr = strings.TrimPrefix(sidesStr, "d")                              //strip d as it is uncessary
     if !isNumeric(sidesStr){                                                  //double-check that remaindes is numeric
         fmt.Println("Error: Problem parsing diceroll")
    }
    sides, _ := strconv.Atoi(sidesStr)                                         //make it an int so it can be parsed


    //Split input on the dice, to get left and right expressions
    splitInput := regexp.MustCompile("d[0-9]+").Split(strings.ToLower(input), 2)


    //Evaluate expressions to left and right of dice
    left, err1 = evaluate(splitInput[0])
    if err1 != nil {
        fmt.Println("Error: Problem with left expression!")
        return 0, err1
    }

    //if the right part is empty, set it to zero to avoid part errors
    fmt.Println(splitInput[1])
    if splitInput[1] == ""{
        splitInput[1] = "0"
    }

    splitInput[1] = "0"+splitInput[1]

    right, err2 = evaluate(splitInput[1])
    if err2 != nil {
        fmt.Println("Error: Problem with right expression!")
        return 0, err2
    }


    diceResults, rollError := RollDice(left, sides)

    if rollError != nil{
        return 0, rollError
    }

    result = sumSlice(diceResults)+int(right.(float64))

    fmt.Printf("%v+%v= ", diceResults, right)

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


func isNumeric (number string) (result bool){
    result, _ = regexp.MatchString("[0-9]+", number)
    return result
}

func sumSlice(slice []int)(result int){
    result = 0
    for i := 0; i< len(slice);i++{
        result += slice[i]
    }
    return result
}