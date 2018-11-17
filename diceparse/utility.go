package diceparse

import "regexp"



//isNumeric takes a string as input and checks if the string is numeric, returning the answer as bool
func isNumeric (number string) (result bool){
    result, _ = regexp.MatchString("[0-9]+", number)
    return result
}

//sumSlice takes a slice of ints as input and returns the sum of all objects in the slice as an int
func sumSlice(slice []int)(result int){
    result = 0
    for i := 0; i< len(slice);i++{
        result += slice[i]
    }
    return result
}

func TrimChar(s *string) string{
    ch := (*s)[:1]
    *s = (*s)[1:]
    return ch
}