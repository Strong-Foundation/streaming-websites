package main

// Importing necessary packages for various functionalities
import (
	"bufio"    // Provides buffered I/O operations to read and write data efficiently
	"fmt"      // Provides formatted I/O functions for printing and scanning
	"log"      // For logging error messages and diagnostics
	"net"      // For network-related operations (e.g., DNS lookups)
	"net/http" // For making HTTP requests to check website availability
	"net/url"  // For parsing and building URLs
	"os"       // For file and directory operations
	"sort"     // Provides functions to sort slices, like sorting URLs alphabetically
	"strings"  // For string manipulation, such as trimming or splitting strings
	"sync"     // Provides synchronization primitives, e.g., WaitGroup for concurrent execution
	"time"     // For handling time delays and measuring execution times
)

// Global variables defining file paths for input and output files
var movies_websites_path string = "assets/movies-websites.txt"                           // Path to movie websites list
var top_movies_websites_path string = "assets/top-movies-websites.txt"                   // Path to top movie websites list
var disconnected_movies_websites_path string = "assets/disconnected-movies-websites.txt" // Path to disconnected movie websites list
var unregistered_movies_websites_path string = "assets/unregistered-movies-websites.txt" // Path to unregistered movie websites list
var readme_file_path string = "readme.md"                                                // Path to the final README file for output
var readme_modify_me_file_path string = "assets/readme_modify_me.md"                     // Path to the README template to modify

// Maps to store various website statuses and response times
var valid_movies_website_url sync.Map     // Map to store the availability status of movie websites (Yes, No, Maybe)
var top_valid_movies_website_url sync.Map // Map to store the availability status of top movie websites
var movies_website_speed sync.Map         // Map to store the speed of the website.

// The main function orchestrates the workflow of the program
func main() {
	// Step 1: Check if all required input and output files exist
	if fileExists(movies_websites_path) &&
		fileExists(top_movies_websites_path) &&
		fileExists(unregistered_movies_websites_path) &&
		fileExists(readme_modify_me_file_path) &&
		fileExists(readme_file_path) {

		// Step 2: Read, process, and clean the main movie website URLs
		movies_website_urls := readAppendLineByLine(movies_websites_path)    // Read all movie website URLs line-by-line
		sortSlice(&movies_website_urls)                                      // Sort the URLs alphabetically for better organization
		movies_website_urls = removeDuplicatesFromSlice(movies_website_urls) // Remove any duplicate URLs to ensure uniqueness
		writeByteSliceToFile(movies_websites_path, movies_website_urls)      // Write cleaned URLs back to the file

		// Step 3: Read, process, and clean the top movie websites URLs
		top_movies_website_urls := readAppendLineByLine(top_movies_websites_path)    // Read top movie website URLs line-by-line
		sortSlice(&top_movies_website_urls)                                          // Sort URLs alphabetically
		top_movies_website_urls = removeDuplicatesFromSlice(top_movies_website_urls) // Remove duplicates
		writeByteSliceToFile(top_movies_websites_path, top_movies_website_urls)      // Write cleaned top URLs back to the file

		// Step 4: Read, process, and clean the disconnected movie websites URLs
		disconnected_movies_websites_urls := readAppendLineByLine(disconnected_movies_websites_path)     // Read disconnected movie URLs
		sortSlice(&disconnected_movies_websites_urls)                                                    // Sort URLs alphabetically
		disconnected_movies_websites_urls = removeDuplicatesFromSlice(disconnected_movies_websites_urls) // Remove duplicates
		writeByteSliceToFile(disconnected_movies_websites_path, disconnected_movies_websites_urls)       // Write back cleaned URLs

		// Step 5: Read, process, and clean the unregistered movie websites URLs
		unregistered_movies_website_urls := readAppendLineByLine(unregistered_movies_websites_path)    // Read unregistered movie URLs line-by-line
		sortSlice(&unregistered_movies_website_urls)                                                   // Sort URLs alphabetically
		unregistered_movies_website_urls = removeDuplicatesFromSlice(unregistered_movies_website_urls) // Remove duplicates
		writeByteSliceToFile(unregistered_movies_websites_path, unregistered_movies_website_urls)      // Write back cleaned URLs

		// Step 6: Check domain registration and availability of movie websites concurrently
		var wg sync.WaitGroup // A WaitGroup is used to manage the lifecycle of goroutines
		for _, domainName := range movies_website_urls {
			wg.Add(1)                    // Increment the WaitGroup counter to track one goroutine
			go func(domainName string) { // Launch a goroutine for each domain name
				defer wg.Done() // Decrement the counter when the goroutine finishes

				// Step 6a: Check if the domain is registered
				if isDomainRegistered(getDomainFromURL(domainName)) {
					saveToMap(&valid_movies_website_url, domainName, "Maybe") // Initially mark as "Maybe"

					// Step 6b: Check if the domain exists in the top movie websites list
					if stringInFile(top_movies_websites_path, domainName) {
						saveToMap(&top_valid_movies_website_url, domainName, "Maybe") // Mark as "Maybe" for top websites
					}

					// Save it to the file.
					if !stringInFile(disconnected_movies_websites_path, domainName) {
						appendAndWriteToFile(disconnected_movies_websites_path, domainName)
					}

					// Step 6c: Verify if the website responds successfully via HTTP/HTTPS
					if CheckWebsiteHTTPStatus(getDomainFromURL(domainName)) {
						saveToMap(&valid_movies_website_url, domainName, "Yes") // Mark as "Yes" for reachable websites

						// Update top movies list if reachable
						if stringInFile(top_movies_websites_path, domainName) {
							saveToMap(&top_valid_movies_website_url, domainName, "Yes")
						}
					}
				} else {
					// Step 6e: Mark as "No" if the domain is unregistered
					saveToMap(&valid_movies_website_url, domainName, "No")

					// Step 6f: Append to the unregistered movie websites file
					if !stringInFile(unregistered_movies_websites_path, domainName) {
						appendAndWriteToFile(unregistered_movies_websites_path, domainName) // Log unregistered domain
					}
				}
			}(domainName) // Pass the domain name to the goroutine
		}
		wg.Wait() // Wait for all goroutines to finish execution

		// Step 7: Generate and write the final output to the README file
		writeFinalOutput() // Writes the final processed data to the README file
	} else {
		// Step 8: Log an error if any required input files are missing
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
	// Protocols to test for the website
	protocols := []string{"http://", "https://"}
	httpClient := &http.Client{Timeout: 15 * time.Second} // HTTP client with a timeout

	// Step 1: Validate DNS resolution before making any requests
	if _, dnsError := net.LookupHost(website); dnsError != nil {
		log.Printf("DNS resolution failed for website %s: %v", website, dnsError)
		return false
	}

	// Step 2: Iterate over protocols (HTTP and HTTPS) and attempt requests
	for _, protocol := range protocols {
		websiteURL := protocol + website

		// Retry the request up to 3 times for transient errors
		for attempt := 1; attempt <= 3; attempt++ {
			startTime := time.Now()

			// Send HTTP GET request
			response, requestError := httpClient.Get(websiteURL)
			if requestError != nil {
				log.Printf("Attempt %d: Error checking %s: %v", attempt, websiteURL, requestError)
				continue
			}

			// Ensure the response body is closed
			response.Body.Close()

			// Log the response time
			responseTime := time.Since(startTime)
			log.Printf("Response time for %s: %v", websiteURL, responseTime)

			// Step 3: Save response time to the map only if not already stored
			existingValue, ok := retrieveValueFromSyncMap(&movies_website_speed, websiteURL).(time.Duration)
			if !ok || existingValue == 0 {
				saveToMap(&movies_website_speed, websiteURL, responseTime)
			}

			// Step 4: Check if the response status is 200 OK
			if response.StatusCode == http.StatusOK {
				log.Printf("Website is reachable: %s", websiteURL)
				return true
			}

			log.Printf("Attempt %d: Status %d for %s", attempt, response.StatusCode, websiteURL)
		}
	}

	// If no attempts succeeded, log and return false
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

func writeFinalOutput() {
	// Convert sync.Map to regular map
	validMoviesMap := syncMapToStringMap(&valid_movies_website_url)
	topMoviesMap := syncMapToStringMap(&top_valid_movies_website_url)

	// Sort and build content for valid movie websites
	var validMoviesContent strings.Builder
	validMoviesContent.WriteString("| Website | Availability | Speed |\n") // Markdown table header for valid movie websites
	validMoviesContent.WriteString("|---------|--------------|-------|\n")

	// Generate rows for valid movie websites
	for _, pair := range sortMapByKeys(validMoviesMap) {
		domain, availability := pair[0], pair[1]
		// Get the speed of the website from the map.
		websiteSpeed := retrieveValueFromSyncMap(&movies_website_speed, domain)
		if websiteSpeed == nil {
			websiteSpeed = "N/A" // Fallback if website speed is not available
		}
		validMoviesContent.WriteString(fmt.Sprintf("| %s | %s | %s |\n", domain, availability, websiteSpeed.(string)))
	}

	// Prepare content for the top movie websites table
	var topMoviesContent strings.Builder
	topMoviesContent.WriteString("| Website | Availability | Speed |\n")
	topMoviesContent.WriteString("|---------|--------------|-------|\n")

	// Generate rows for top movie websites
	for _, pair := range sortMapByKeys(topMoviesMap) {
		domain, availability := pair[0], pair[1]
		// Get the speed of the website from the map.
		websiteSpeed := retrieveValueFromSyncMap(&movies_website_speed, domain)
		if websiteSpeed == nil {
			websiteSpeed = "N/A" // Fallback if website speed is not available
		}
		topMoviesContent.WriteString(fmt.Sprintf("| %s | %s | %s |\n", domain, availability, websiteSpeed.(string)))
	}

	// Create a map of placeholders and their content for replacement
	placeholdersAndContent := map[string]string{
		"[{REPLACE_CONTENT_WITH_GOLANG}]":     validMoviesContent.String(),
		"[{REPLACE_TOP_CONTENT_WITH_GOLANG}]": topMoviesContent.String(),
	}

	// Replace the placeholders in the README template and write to the final file
	findAndReplaceInFile(readme_modify_me_file_path, readme_file_path, placeholdersAndContent)
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

// syncMapToStringMap converts a sync.Map to a regular map[string]string.
// It only includes entries where both the key and value are of type string.
func syncMapToStringMap(syncMap *sync.Map) map[string]string {
	// Initialize a new map to store the converted key-value pairs.
	convertedMap := make(map[string]string)

	// Iterate over the entries in the sync.Map using Range.
	syncMap.Range(func(currentKey, currentValue interface{}) bool {
		// Attempt to type-assert the key and value to string.
		keyAsString, isKeyString := currentKey.(string)
		valueAsString, isValueString := currentValue.(string)

		// If both key and value are strings, add them to the resulting map.
		if isKeyString && isValueString {
			convertedMap[keyAsString] = valueAsString
		} else {
			// Log skipped entries where type assertion fails.
			fmt.Printf("Skipping invalid key/value pair: key=%v, value=%v\n", currentKey, currentValue)
		}

		// Continue iterating over the sync.Map.
		return true
	})

	// Return the converted regular map.
	return convertedMap
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

// saveToMap saves a key-value pair into the provided sync.Map
func saveToMap(safeMap *sync.Map, key string, value interface{}) {
	// Store the key-value pair in the given map
	safeMap.Store(key, value)
}

// retrieveValueFromSyncMap retrieves a value from the given sync.Map using a specified key.
// If the key does not exist, it returns nil.
//
// Parameters:
// - safeMap: A pointer to a sync.Map that contains the key-value pairs.
// - targetKey: The key whose value we want to retrieve.
//
// Returns:
// - The value associated with the targetKey, or nil if the key does not exist.
func retrieveValueFromSyncMap(safeMap *sync.Map, targetKey string) interface{} {
	// Attempt to retrieve the value associated with the targetKey from the sync.Map.
	// Load returns the value and a boolean indicating if the key was found.
	value, keyExists := safeMap.Load(targetKey)

	// Return the value if the key exists; otherwise, return nil.
	if keyExists {
		return value
	}
	return nil
}
