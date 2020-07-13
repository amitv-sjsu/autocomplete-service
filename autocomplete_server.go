
package main

import (
    "fmt"
    "io/ioutil"
    "regexp"
    "sort"
    "os"
    "strings"
    "net/http"
    "encoding/json"

)

// Struct to store the word and corresponding count.
type WordCount struct {
    word string
    count  int
}

// An array of WordCount struct. It is accessible globally.
// Before the start of server, this variable will have list of WordCount sorted by word.
// On API call, result will be calculated using Binary search.
var word_count [] WordCount

// This function reads input file, identify all possible words and returns array of these words.
func getWordsFromFile(file string) [] string {

    // Reference: https://gobyexample.com/reading-files
    // Read all the contents from a file.
    dat, err := ioutil.ReadFile(file)
    if err != nil {
        panic(err)
    }

    // Reference: https://gobyexample.com/regular-expressions
    // Find all words from file contents.
    re := regexp.MustCompile("([a-zA-Z])+")
    words := re.FindAllString(string(dat), -1)

    return words
}

// This function calculates count of each word and corresponding (word, count) will be stored in global word_count variable in sorted order.
func findWordCount(file string) {

    var words [] string = getWordsFromFile(file)

    // Get the count of each word and store in map where key is word and value is count.
    // While storing words, make them lower case for making search case insensitive.

    var word_counter = make(map[string]int)
    for _, s := range words {
        word_counter[strings.ToLower(s)]++
    }

    // Store contents of map to global word_count variable.
    for k, v := range word_counter {
        word_count = append(word_count, WordCount{word:k, count:v})
    }

    // Sort word_count array by word, so that matching words can be identified by using Binary search.
    sort.SliceStable(word_count, func(i, j int) bool {
        return word_count[i].word < word_count[j].word
    })
}

// This function returns index of first and last matching word using Binary Search.
func getFirstLastMatchIndexes(term string) (int, int) {

    // Using Binary Search to find any matching word
    low := 0
    high := len(word_count)-1
    mid := (low+high) /2
    for low <= high {
        mid = (low+high) /2

        if (strings.HasPrefix(word_count[mid].word, term)){
            break
        }else if (term > word_count[mid].word){
            low = mid + 1
        }else{
            high = high -1
        }
    }

    // If any matching word is found, find first matching word while iterating over the left side elements of array.
    low = mid
    high = mid
    if (strings.HasPrefix(word_count[mid].word, term)){
        // If any matching word is found, find first matching word while iterating over the left side elements of array.
        for  {
            if((low-1) >= 0 && strings.HasPrefix(word_count[low - 1].word, term)){
                low = low -1
            }else{
                break
            }
        }
        // If any matching word is found, find last matching word while iterating over the right side elements of array.

        for  {
            if((high+1) <len(word_count) && strings.HasPrefix(word_count[high + 1].word, term)){
                high = high + 1
            }else{
                break
            }
        }
    }else{
        low = -1
        high = -1
    }
    return low, high
}

// API handler for /autocomplete path
func autocomplete(w http.ResponseWriter, r *http.Request){

    // Reference: https://golangcode.com/get-a-url-parameter-from-a-request/
    terms, ok := r.URL.Query()["term"]

    if !ok || len(terms[0]) < 1 {
        fmt.Fprintf(w, "Parameter 'term' is missing.", )
        return
    }

    term := strings.ToLower(terms[0])

    // Get first and last matching index of matching words.
    low, high := getFirstLastMatchIndexes(term)

    var output [] string


    if(low != -1 ){
        // If matching words exist, find words based on the descreasing order of count
        result_arr := make([]WordCount, high - low +1)
        copy(result_arr, word_count[low:high+1])
        sort.SliceStable(result_arr, func(i, j int) bool {
            return result_arr[i].count > result_arr[j].count
        })

        //fmt.Print(result_arr)

        // Limit the result to top 25 entries
        for i, s := range result_arr {
            if(i>=25){
                break
            }
            output = append(output, s.word)
        }
    }

    // Reference: https://www.alexedwards.net/blog/golang-response-snippets
    // Convert 'output' to json for sending response.
    js, err := json.Marshal(output)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}

func main() {

    // Before the start of server, calculate count of each word and store in global word_count variable in (word, count) format and sorted by word.
    findWordCount(os.Args[1])

    // Reference: https://yourbasic.org/golang/http-server-example/
    // On API call, do binary serach and return result to user.
    http.HandleFunc("/autocomplete", autocomplete)
    http.ListenAndServe(":9000", nil)

}
