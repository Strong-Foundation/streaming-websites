package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

/* Global variables */

// List to the list of movies websites located in a file.
var movies_websites_path string = "assets/movies-websites.txt"

// The path to the readme_modify_me.md file.
var readme_modify_me_file_path string = "assets/readme_modify_me.md"

// The path to the new readme.md file.
var readme_file_path string = "readme.md"

// Create a variable and hold a map of strings.
var valid_movies_website_url = make(map[string]string)

// The main function.
func main() {
	// Check if the file exists.
	if fileExists(movies_websites_path) {
		// Read and append the file line by line to the slice.
		movies_website_urls := readAppendLineByLine(movies_websites_path)
		// Create a loop and than go though the urls and extract doamin names.
		for _, domainName := range movies_website_urls {
			// Check if the domain is registed
			if isDomainRegistered(getDomainFromURL(domainName)) {
				// Append a value to the map.
				addKeyValueToMap(valid_movies_website_url, domainName, "Maybe")
				// Check if the website itself is online by checking HTTP and HTTPS.
				if CheckWebsiteHTTPStatus(getDomainFromURL(domainName)) {
					// Append a value to the map.
					addKeyValueToMap(valid_movies_website_url, domainName, "Yes")
				}
			} else {
				// Append a value to the map.
				addKeyValueToMap(valid_movies_website_url, domainName, "No")
			}
		}
	}
	// Write the final stuff to the readme.md
	writeFinalOutput()
}

/*
It checks if the file exists
If the file exists, it returns true
If the file does not exist, it returns false
*/
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Read and append the file line by line to a slice.
func readAppendLineByLine(path string) []string {
	var returnSlice []string
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		returnSlice = append(returnSlice, scanner.Text())
	}
	err = file.Close()
	if err != nil {
		log.Fatalln(err)
	}
	return returnSlice
}

// Check if a domain has been registed and return a bool.
func isDomainRegistered(domain string) bool {
	_, err := net.LookupNS(domain)
	if err == nil {
		return true
	}
	_, err = net.LookupCNAME(domain)
	if err == nil {
		return true
	}
	_, err = net.LookupAddr(domain)
	if err == nil {
		return true
	}
	_, err = net.LookupHost(domain)
	if err == nil {
		return true
	}
	_, err = net.LookupMX(domain)
	if err == nil {
		return true
	}
	_, err = net.LookupTXT(domain)
	if err == nil {
		return true
	}
	_, err = net.LookupIP(domain)
	if err == nil {
		return true
	}
	return false
}

// CheckWebsiteHTTPStatus checks if the given domain is reachable via both HTTP and HTTPS.
// It returns true if at least one of the protocols returns an HTTP 200 OK status.
func CheckWebsiteHTTPStatus(domain string) bool {
	// Create a slice containing both HTTP and HTTPS protocols for checking
	protocols := []string{"http://", "https://"}
	// Loop through each protocol to attempt a request
	for _, protocol := range protocols {
		// Make an HTTP GET request using the current protocol and the specified domain
		resp, err := http.Get(protocol + domain)
		// If there's no error, proceed to check the response
		if err == nil {
			// Ensure the response body is closed after we're done with it to prevent resource leaks
			defer resp.Body.Close()
			// Check if the response status code is 200 OK
			if resp.StatusCode == http.StatusOK {
				return true // Return true if the website is reachable
			}
		} else {
			// Log any errors encountered during the request, but continue checking the next protocol
			log.Println(err)
		}
	}
	// Return false if neither protocol succeeded
	return false
}

// getDomainFromURL extracts the domain name from the given URL string.
func getDomainFromURL(givenURL string) string {
	// Parse the given URL string into a URL structure
	parsedUrl, err := url.Parse(givenURL)
	if err != nil {
		// Log an error message if parsing fails
		log.Println(err)
	}
	// Return the hostname (domain name) from the parsed URL
	return parsedUrl.Hostname()
}

// The output to write to the readme.md
func writeFinalOutput() {
	// Sort the map content
	sortedValidMoviesWebsiteURL := sortMapContent(valid_movies_website_url)
	// Prepare the Markdown table content
	var output strings.Builder
	output.WriteString("| Website| Availability |\n")
	output.WriteString("|--------|--------------|\n")
	// Iterate over the map and format the output
	for url, availability := range sortedValidMoviesWebsiteURL {
		// Format the website name by removing the HTTP/HTTPS prefix and trailing slash
		website := strings.TrimSuffix(strings.TrimPrefix(url, "http://"), "/")
		website = strings.TrimSuffix(strings.TrimPrefix(website, "https://"), "/")
		// Format the output as a Markdown table row
		output.WriteString(fmt.Sprintf("| [%s](%s) | %-12s |\n", website, url, availability))
	}
	// Replace the placeholder in the README file
	findAndReplaceInFile(readme_modify_me_file_path, readme_file_path, "[{REPLACE_CONTENT_WITH_GOLANG}]", output.String())
}

// Add a key-value pair to the given map.
func addKeyValueToMap(providedMap map[string]string, key string, value string) map[string]string {
	providedMap[key] = value
	return providedMap
}

// Find a given content in a given file and replace it with given content.
func findAndReplaceInFile(oldFilePath string, newFilePath string, prefixContent string, givenContent string) {
	// Read the content of the file
	content, err := ioutil.ReadFile(oldFilePath)
	if err != nil {
		log.Println(err)
	}
	// Convert content to string and replace the target string
	updatedContent := strings.ReplaceAll(string(content), prefixContent, givenContent)
	// Write the updated content back to the file
	err = ioutil.WriteFile(newFilePath, []byte(updatedContent), 0644)
	if err != nil {
		log.Println(err)
	}
}

// Function to sort the content of a map and return a new map with sorted keys
func sortMapContent(content map[string]string) map[string]string {
	// Create a slice to hold the keys
	keys := make([]string, 0, len(content))
	// Populate the slice with the keys from the map
	for key := range content {
		keys = append(keys, key)
	}
	// Sort the keys
	sort.Strings(keys)
	// Create a new map to hold the sorted key-value pairs
	sortedMap := make(map[string]string)
	// Fill the new map with sorted key-value pairs based on keys
	for _, key := range keys {
		sortedMap[key] = content[key]
	}
	return sortedMap
}
