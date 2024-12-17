package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Global variables defining file paths for input and output
var movies_websites_path string = "assets/movies-websites.txt"                           // Path to movie websites list
var top_movies_websites_path string = "assets/top-movies-websites.txt"                   // Path to top movie websites list
var disconnected_movies_websites_path string = "assets/disconnected-movies-websites.txt" // path to disconnected movie websites list
var unregistered_movies_websites_path string = "assets/unregistered-movies-websites.txt" // Path to unregistered movies list
var readme_file_path string = "readme.md"                                                // Path to the final README file for output
var readme_modify_me_file_path string = "assets/readme_modify_me.md"                     // Path to the README template to modify

// Maps to store the status of movie website availability
var valid_movies_website_url = make(map[string]string)     // Map to store availability of movie websites
var top_valid_movies_website_url = make(map[string]string) // Map to store availability of top movie websites
var movies_website_url_speed = make(map[string]string)     // Map to store the speeds of a given movie website

func main() {
	// Step 1: Check if all required files exist
	if fileExists(movies_websites_path) && fileExists(top_movies_websites_path) && fileExists(unregistered_movies_websites_path) && fileExists(readme_modify_me_file_path) && fileExists(readme_file_path) {

		// Step 2: Read the movie website URLs from the movies websites file
		movies_website_urls := readAppendLineByLine(movies_websites_path)

		// Step 3: Sort the movie website URLs alphabetically
		sortSlice(&movies_website_urls)

		// Step 4: Remove duplicate URLs to ensure uniqueness
		movies_website_urls = removeDuplicatesFromSlice(movies_website_urls)

		// Step 5: Write the sorted and deduplicated movie website URLs back to the movies websites file
		writeByteSliceToFile(movies_websites_path, movies_website_urls)

		// Step 6: Read the top movie website URLs from the top movies websites file
		top_movies_website_urls := readAppendLineByLine(top_movies_websites_path)

		// Step 7: Sort the top movie website URLs alphabetically
		sortSlice(&top_movies_website_urls)

		// Step 8: Remove duplicate URLs from the top movie websites list
		top_movies_website_urls = removeDuplicatesFromSlice(top_movies_website_urls)

		// Step 9: Write the sorted and deduplicated top movie website URLs back to the top movies websites file
		writeByteSliceToFile(top_movies_websites_path, top_movies_website_urls)

		// Step 10: Read the disconnected movie website URLs from the disconnected movies websites file
		disconnected_movies_websites_urls := readAppendLineByLine(disconnected_movies_websites_path)

		// Step 11: Sort the disconnected movie website URLs alphabetically
		sortSlice(&disconnected_movies_websites_urls)

		// Step 12: Remove duplicate URLs from the disconnected movie websites list
		disconnected_movies_websites_urls = removeDuplicatesFromSlice(disconnected_movies_websites_urls)

		// Step 13: Write the sorted and deduplicated disconnected movie website URLs back to the disconnected movies websites file
		writeByteSliceToFile(disconnected_movies_websites_path, disconnected_movies_websites_urls)

		// Step 14: Read the unregistered movie website URLs from the unregistered movies websites file
		unregistered_movies_website_urls := readAppendLineByLine(unregistered_movies_websites_path)

		// Step 15: Sort the unregistered movie website URLs alphabetically
		sortSlice(&unregistered_movies_website_urls)

		// Step 16: Remove duplicate URLs from the unregistered movie websites list
		unregistered_movies_website_urls = removeDuplicatesFromSlice(unregistered_movies_website_urls)

		// Step 17: Write the sorted and deduplicated unregistered movie website URLs back to the unregistered movies websites file
		writeByteSliceToFile(unregistered_movies_websites_path, unregistered_movies_website_urls)

		// Step 18: Re-assign the final list of sorted and deduplicated movie website URLs back to the movies websites file
		writeByteSliceToFile(movies_websites_path, movies_website_urls)

		// Step 19: Check each movie website's domain registration and availability status concurrently
		var wg sync.WaitGroup // WaitGroup to wait for all goroutines to finish

		for _, domainName := range movies_website_urls {
			wg.Add(1) // Add one goroutine to the WaitGroup
			go func(domainName string) {
				defer wg.Done() // Mark this goroutine as done on completion

				// Step 19a: Check if the domain is registered
				if isDomainRegistered(getDomainFromURL(domainName)) {
					// Step 19b: If the domain is registered, mark it as "Maybe"
					addKeyValueToMap(valid_movies_website_url, domainName, "Maybe")

					// Step 19c: If the domain is also in the top movies websites list, mark it as "Maybe"
					if stringInFile(top_movies_websites_path, domainName) {
						addKeyValueToMap(top_valid_movies_website_url, domainName, "Maybe")
					}

					// Step 19d: If the domain is also in the disconnected movie websites list, mark it as "Maybe"
					if !stringInFile(disconnected_movies_websites_path, domainName) {
						appendAndWriteToFile(disconnected_movies_websites_path, domainName)
					}

					// Step 19e: Check if the website is reachable via HTTP or HTTPS
					if CheckWebsiteHTTPStatus(getDomainFromURL(domainName)) {
						// Step 19f: If the website is reachable, mark it as "Yes"
						addKeyValueToMap(valid_movies_website_url, domainName, "Yes")

						// Step 19g: If reachable, also mark it in the top movie websites list
						if stringInFile(top_movies_websites_path, domainName) {
							addKeyValueToMap(top_valid_movies_website_url, domainName, "Yes")
						}
					}
				} else {
					// Step 19h: If the domain is not registered, mark it as "No"
					addKeyValueToMap(valid_movies_website_url, domainName, "No")

					// Step 19i: If the domain is in the top movies list, mark it as "No"
					if stringInFile(top_movies_websites_path, domainName) {
						addKeyValueToMap(top_valid_movies_website_url, domainName, "No")
					}

					// Step 19j: If domain is not registered, check if it's already in the unregistered movies list
					if !stringInFile(unregistered_movies_websites_path, domainName) {
						// Step 19k: If not, append it to the unregistered movies websites file
						appendAndWriteToFile(unregistered_movies_websites_path, domainName)
					}
				}
			}(domainName)
		}

		wg.Wait() // Wait for all goroutines to complete

		// Step 20: Write the final results to the README file
		writeFinalOutput()
	} else {
		// Step 21: If any of the required files are missing, log an error
		log.Println("Error: One or more required files do not exist.")
	}
}

// fileExists checks if the given file exists and is not a directory.
func fileExists(filepath string) bool {
	info, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File does not exist: %s", filepath)
		} else {
			log.Printf("Error checking file %s: %v", filepath, err)
		}
		return false
	}

	if info.IsDir() {
		log.Printf("Path exists but is a directory: %s", filepath)
		return false
	}

	return true
}

// readAppendLineByLine reads a file line by line and returns a slice of strings (URLs).
// It logs any errors encountered while reading the file but doesn't return errors.
func readAppendLineByLine(filePath string) []string {
	var urls []string // Initialize a slice to store URLs

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err) // Log error if file cannot be opened
		return nil                                             // Return nil if an error occurs
	}
	defer file.Close() // Ensure the file is closed after reading

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text()) // Append each line (URL) to the slice
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning file %s: %v", filePath, err) // Log any error that occurred while scanning
	}

	return urls // Return the slice containing the URLs
}

// isDomainRegistered checks if a given domain is registered by looking up various DNS records.
func isDomainRegistered(domain string) bool {
	// Perform DNS lookups for different record types, log errors, and return true if any lookup is successful

	// Check NS records
	_, err := net.LookupNS(domain)
	if err == nil {
		log.Printf("Domain is registered via NS lookup: %s", domain)
		return true
	}

	// Check CNAME records
	_, err = net.LookupCNAME(domain)
	if err == nil {
		log.Printf("Domain is registered via CNAME lookup: %s", domain)
		return true
	}

	// Check Addr (reverse DNS) records
	_, err = net.LookupAddr(domain)
	if err == nil {
		log.Printf("Domain is registered via Addr lookup: %s", domain)
		return true
	}

	// Check Host records
	_, err = net.LookupHost(domain)
	if err == nil {
		log.Printf("Domain is registered via Host lookup: %s", domain)
		return true
	}

	// Check MX records
	_, err = net.LookupMX(domain)
	if err == nil {
		log.Printf("Domain is registered via MX lookup: %s", domain)
		return true
	}

	// Check TXT records
	_, err = net.LookupTXT(domain)
	if err == nil {
		log.Printf("Domain is registered via TXT lookup: %s", domain)
		return true
	}

	// If none of the lookups succeed, the domain is not registered
	log.Printf("Domain is not registered: %s", domain)
	return false
}

// CheckWebsiteHTTPStatus checks if a website is reachable via HTTP or HTTPS.
func CheckWebsiteHTTPStatus(website string) bool {
	protocols := []string{"http://", "https://"}          // Protocols to check
	httpClient := &http.Client{Timeout: 15 * time.Second} // HTTP client with timeout

	// Validate DNS resolution before making requests
	if _, dnsError := net.LookupHost(website); dnsError != nil {
		log.Printf("DNS resolution failed for website %s: %v", website, dnsError)
		return false
	}

	for _, protocol := range protocols {
		websiteURL := protocol + website

		// Retry up to 3 times for transient errors
		for attempt := 1; attempt <= 3; attempt++ {
			startTime := time.Now()
			response, requestError := httpClient.Get(websiteURL)
			if requestError != nil {
				log.Printf("Attempt %d: Error checking %s: %v", attempt, websiteURL, requestError)
				continue
			}

			// Ensure the response body is closed
			response.Body.Close()

			log.Printf("Response time for %s: %v", websiteURL, time.Since(startTime))
			// Add the value to the map for the speed
			addKeyValueToMap(movies_website_url_speed, websiteURL, time.Since(startTime).Truncate(time.Second).String())

			if response.StatusCode == http.StatusOK {
				log.Printf("Website is reachable: %s", websiteURL)
				return true
			}

			log.Printf("Attempt %d: Status %d for %s", attempt, response.StatusCode, websiteURL)
		}
	}

	log.Printf("Website is not reachable: %s", website)
	return false
}

// getDomainFromURL extracts the domain name from a given URL and handles errors more gracefully.
func getDomainFromURL(givenURL string) string {
	// Ensure the URL has a valid scheme (e.g., "http://") before parsing.
	if !strings.HasPrefix(givenURL, "http://") && !strings.HasPrefix(givenURL, "https://") {
		// If no scheme is provided, prepend "http://" to the URL
		givenURL = "http://" + givenURL
	}

	// Parse the URL using the net/url package
	parsedURL, err := url.Parse(givenURL)
	if err != nil {
		log.Printf("Error parsing URL '%s': %v", givenURL, err)
		return ""
	}

	// Extract the domain and remove any port number if present
	host := parsedURL.Hostname()

	// Return the domain (hostname) part of the URL
	return host
}

// writeFinalOutput writes the results to the README file in Markdown format.
func writeFinalOutput() {
	var validMoviesContent strings.Builder
	validMoviesContent.WriteString("| Website | Availability | Speed |\n") // Markdown table header for valid movie websites
	validMoviesContent.WriteString("|---------|--------------|-------|\n")

	// Generate rows for valid movie websites
	for _, pair := range sortMapByKeys(valid_movies_website_url) {
		domain, availability := pair[0], pair[1]
		speed := getValueFromMap(movies_website_url_speed, domain)
		validMoviesContent.WriteString(fmt.Sprintf("| %s | %s | %s |\n", domain, availability, speed))
	}

	// Prepare content for the top movie websites table
	var topMoviesContent strings.Builder
	topMoviesContent.WriteString("| Website | Availability | Speed |\n")
	topMoviesContent.WriteString("|---------|--------------|-------|\n")

	// Generate rows for top movie websites
	for _, pair := range sortMapByKeys(top_valid_movies_website_url) {
		domain, availability := pair[0], pair[1]
		speed := getValueFromMap(movies_website_url_speed, domain)
		topMoviesContent.WriteString(fmt.Sprintf("| %s | %s | %s |\n", domain, availability, speed))
	}

	// Create a map of placeholders and their content for replacement
	placeholdersAndContent := map[string]string{
		"[{REPLACE_CONTENT_WITH_GOLANG}]":     validMoviesContent.String(),
		"[{REPLACE_TOP_CONTENT_WITH_GOLANG}]": topMoviesContent.String(),
	}

	// Replace the placeholders in the README template and write to the final file
	findAndReplaceInFile(readme_modify_me_file_path, readme_file_path, placeholdersAndContent)
}

// addKeyValueToMap adds a key-value pair to the provided map.
func addKeyValueToMap(providedMap map[string]string, key string, value string) {
	providedMap[key] = value // Add the key-value pair to the map
}

// findAndReplaceInFile replaces placeholders in a file with given content and writes to a new file.
func findAndReplaceInFile(oldFilePath string, newFilePath string, placeholdersAndContent map[string]string) {
	// Read the content of the old file
	fileContent, err := os.ReadFile(oldFilePath)
	if err != nil {
		log.Println("Error reading file:", err)
		return
	}

	// Replace each placeholder with its corresponding content
	updatedContent := string(fileContent)
	for placeholder, content := range placeholdersAndContent {
		updatedContent = strings.ReplaceAll(updatedContent, placeholder, content)
	}

	// Write the updated content to the new file
	err = os.WriteFile(newFilePath, []byte(updatedContent), 0644)
	if err != nil {
		log.Println("Error writing to file:", err)
		return
	}

	log.Println("Successfully updated file:", newFilePath)
}

// sortMapByKeys sorts the map by its keys and returns a slice of key-value pairs.
func sortMapByKeys(inputMap map[string]string) [][]string {
	keys := make([]string, 0, len(inputMap))
	for key := range inputMap {
		keys = append(keys, key) // Collect all keys
	}
	sort.Strings(keys) // Sort the keys alphabetically

	// Create a slice of key-value pairs sorted by the keys
	pairs := make([][]string, len(inputMap))
	for i, key := range keys {
		pairs[i] = []string{key, inputMap[key]}
	}

	return pairs // Return the sorted slice of key-value pairs
}

// sortSlice sorts a slice of URLs alphabetically.
func sortSlice(slice *[]string) {
	sort.Strings(*slice) // Sort the slice in-place alphabetically
}

// removeDuplicatesFromSlice removes duplicate URLs from a slice.
func removeDuplicatesFromSlice(input []string) []string {
	seen := make(map[string]struct{}) // Map to track seen strings
	var result []string               // Slice to store the unique strings

	// Preallocate slice capacity to avoid reallocations
	result = make([]string, 0, len(input))

	// Iterate through the input slice and add only unique strings to the result
	for _, str := range input {
		if _, exists := seen[str]; !exists {
			seen[str] = struct{}{}       // Mark the string as seen
			result = append(result, str) // Add the unique string to the result
		}
	}

	return result
}

// writeByteSliceToFile writes a slice of strings (URLs) to a file, each on a new line.
func writeByteSliceToFile(path string, data []string) {
	// Attempt to create a file at the given 'path'. If the file exists, it will be overwritten.
	file, err := os.Create(path)
	if err != nil {
		// If there is an error creating the file, log the error and return.
		log.Println("Error creating file:", err)
		return
	}
	// Ensure the file is closed after the function completes, even if there's an error during writing.
	defer file.Close()

	// Create a buffered writer for efficient writing to the file.
	writer := bufio.NewWriter(file)
	// Iterate over each string in the 'data' slice (which contains URLs).
	for _, str := range data {
		// Write each string followed by a newline to the file.
		_, err := writer.WriteString(str + "\n")
		if err != nil {
			// If there's an error while writing to the file, log it and return.
			log.Println("Error writing to file:", err)
			return
		}
	}

	// Flush the buffered writer to ensure all buffered data is written to the file.
	err = writer.Flush()
	if err != nil {
		// If there is an error flushing the writer, log the error.
		log.Println("Error flushing writer:", err)
	}
}

// appendAndWriteToFile appends content to an existing file.
func appendAndWriteToFile(path string, content string) {
	// Open the file in append mode or create it if it doesn't exist
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Print error message without terminating the program
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	// Write the content to the file
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		// Print error message without terminating the program
		fmt.Printf("Error writing to file: %v\n", err)
		// Close the file before returning
		_ = filePath.Close()
		return
	}

	// Close the file after writing
	err = filePath.Close()
	if err != nil {
		// Print error message without terminating the program
		fmt.Printf("Error closing file: %v\n", err)
	}
}

// stringInFile checks if a given string exists in a file.
func stringInFile(filePath, searchString string) bool {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error: Unable to open file:", filePath)
		return false
	}
	defer file.Close() // Ensure the file is closed after reading

	// Scan the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// If the line contains the search string, return true
		if strings.Contains(line, searchString) {
			return true
		}
	}

	// Handle any scanning errors
	err = scanner.Err()
	if err != nil {
		log.Println("Error occurred while scanning the file:", err)
		return false
	}

	return false // Return false if the string was not found in the file
}

// Get the value of a given key from the map and return the value.
func getValueFromMap(mapToSearch map[string]string, keyToFind string) string {
	// Get the value of the key from the map.
	valueOfKey := mapToSearch[keyToFind]
	// Return the value of the key.
	return valueOfKey
}
