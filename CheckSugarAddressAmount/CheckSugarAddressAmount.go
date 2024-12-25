package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Use the api URL
var apiServer = "https://api.sugar.wtf/balance/"

// Exported wallet address file
var TxtFileaddr = "sugar1qnayg20edjqshzd9cuzungycygejdujqacpx9y0.txt"

// The address with the balance is saved to the file location
var outFile = "output.txt"

// Match the characters between the two strings
var startStr = "addr="
var endStr = " "

func main() {
	// Extract all matching strings
	allExtractedStrings, err := extractAllBetweenFromFile(TxtFileaddr, startStr, endStr)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Get wallet address
	for i, str := range allExtractedStrings {
		//fmt.Printf("Extracted string %d: %s\n", i+1, str)
		jsondata, err := GetJsonBlance(getBlance(str))
		if err != nil {
			for err != nil {
				fmt.Println(i, str, "err code 502,retry in 500ms")
				time.Sleep(500 * time.Millisecond)
				jsondata, err = GetJsonBlance(getBlance(str))
			}
		}
		fmt.Println(i, str, "amount=", jsondata)
		if jsondata != 0 {
			fmt.Println("find it--iniyou--", i, str)
			// WriteStringToFile Save to outFile
			err := WriteStringToFile(outFile, str+"    amount: "+string(jsondata))
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Successfully wrote to the file.")
			}
		}

	}

}

func getBlance(addr string) string {
	// create HTTP Client
	client := &http.Client{}

	// create request
	req, err := http.NewRequest("GET", apiServer+addr, nil)
	if err != nil {
		panic(err)
	}

	// Send request to ApiServer
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close() // Make sure the responder body is closed

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func extractAllBetweenFromFile(filePath, startStr, endStr string) ([]string, error) {
	// Read The filePath file contents
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	content := string(data)

	var results []string
	startIndex := 0

	for {
		// Find the next starting position
		startPos := strings.Index(content[startIndex:], startStr)
		if startPos == -1 {
			break // There are no more starting tags
		}
		startPos += startIndex + len(startStr) // Adjust to absolute position and skip the starting flag

		// Find the next end location
		endPos := strings.Index(content[startPos:], endStr)
		if endPos == -1 {
			break // There is no corresponding end flag
		}
		endPos += startPos // Adjust to absolute position

		// Extract and save the substring
		extractedStr := content[startPos:endPos]

		// The string after the newline character is removed
		// Creates a regular expression pattern that matches a newline and everything that follows it
		re := regexp.MustCompile(`\n.*`)
		re2 := regexp.MustCompile(`\r.*`)

		// Use the ReplaceAllString method to remove the matched part
		extractedStr = re.ReplaceAllString(extractedStr, "")
		extractedStr = re2.ReplaceAllString(extractedStr, "")

		results = append(results, extractedStr)

		// Update the search starting point to after the current end tag
		startIndex = endPos + len(endStr)
	}
	return results, nil
}

// Define a structure that matches the JSON data structure
type Response struct {
	Error  *string    `json:"error"` // Use pointer types to handle null
	ID     string     `json:"id"`
	Result ResultData `json:"result"`
}

type ResultData struct {
	Balance  uint64 `json:"balance"`  // Choose the right type for your data
	Received uint64 `json:"received"` // Other fields are also included here if you need them
}

func GetJsonBlance(jsonString string) (uint64, error) {
	// Create a Response variable to store the decoded data
	var response Response
	// Decode the JSON string into response struct
	err := json.Unmarshal([]byte(jsonString), &response)
	if err != nil {
		//fmt.Println("err=", err)
		return 0, err
	}
	// Check for errors
	if response.Error != nil {
		//fmt.Printf("Error occurred: %s\n", *response.Error)
		return 0, err
	}
	// return Balance
	return response.Result.Balance, err
}

// WriteStringToFile,Writes the given string contents to the specified file.
// If the file does not exist, it is created; If the file exists, the content is appended to the end of the file.
func WriteStringToFile(filePath string, content string) error {
	var file *os.File
	var err error

	// Check whether the file exists
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		// The file does not exist. Create a new file and open it in write mode
		file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	} else {
		// File exists, opened in append mode
		file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	}

	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close() // Be sure to close the file at the end of the function

	// Writes a string to a file
	_, err = fmt.Fprintln(file, content) // Use Fprintln to add a newline at the end
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil // No error occurred
}
