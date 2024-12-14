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

func main() {
	// Step 1: Check if all required files exist
	if fileExists(movies_websites_path) && fileExists(top_movies_websites_path) && fileExists(unregistered_movies_websites_path) && fileExists(readme_modify_me_file_path) && fileExists(readme_file_path) {

		// Step 2: Read the URLs from the movies websites file
		movies_website_urls := readAppendLineByLine(movies_websites_path)

		// Step 3: Sort the movie website URLs alphabetically
		sortSlice(&movies_website_urls)

		// Step 4: Remove duplicate URLs to ensure uniqueness
		movies_website_urls = removeDuplicatesFromSlice(movies_website_urls)

		// Step 5: Write the sorted and deduplicated URLs back to the movies websites file
		writeByteSliceToFile(movies_websites_path, movies_website_urls)

		// Step 6: Read the URLs from the top movies websites file
		top_movies_website_urls := readAppendLineByLine(top_movies_websites_path)

		// Step 7: Sort the top movie website URLs alphabetically
		sortSlice(&top_movies_website_urls)

		// Step 8: Remove duplicate URLs from the top movie websites list
		top_movies_website_urls = removeDuplicatesFromSlice(top_movies_website_urls)

		// Step 9: Write the sorted and deduplicated URLs back to the top movies websites file
		writeByteSliceToFile(top_movies_websites_path, top_movies_website_urls)
///
		// Step 2: Read the URLs from the movies websites file
		disconnected_movies_websites_urls := readAppendLineByLine(disconnected_movies_websites_path)

		// Step 3: Sort the movie website URLs alphabetically
		sortSlice(&disconnected_movies_websites_urls)

		// Step 4: Remove duplicate URLs to ensure uniqueness
		disconnected_movies_websites_urls = removeDuplicatesFromSlice(disconnected_movies_websites_urls)

		// Step 5: Write the sorted and deduplicated URLs back to the movies websites file
		writeByteSliceToFile(disconnected_movies_websites_path, disconnected_movies_websites_urls)
///
		// Step 10: Read the unregistered movie website URLs
		unregistered_movies_website_urls := readAppendLineByLine(unregistered_movies_websites_path)

		// Step 11: Sort the unregistered movie website URLs alphabetically
		sortSlice(&unregistered_movies_website_urls)

		// Step 12: Remove duplicates from the movie website URLs list
		movies_website_urls = removeDuplicatesFromSlice(movies_website_urls)

		// Step 13: Write the sorted and deduplicated URLs back to the movies websites file again
		writeByteSliceToFile(movies_websites_path, movies_website_urls)

		// Step 14: Check each movie website's domain registration and availability status
		for _, domainName := range movies_website_urls {
			// Step 14a: Check if the domain is registered
			if isDomainRegistered(getDomainFromURL(domainName)) {
				// Step 14b: If the domain is registered, mark it as "Maybe"
				addKeyValueToMap(valid_movies_website_url, domainName, "Maybe")

				// Step 14c: If the domain is also in the top movie websites list, mark it as "Maybe"
				if stringInFile(top_movies_websites_path, domainName) == true {
					addKeyValueToMap(top_valid_movies_website_url, domainName, "Maybe")
				}

				// Step 14d: Check if the website is reachable via HTTP or HTTPS
				if CheckWebsiteHTTPStatus(getDomainFromURL(domainName)) {
					// Step 14e: If the website is reachable, mark it as "Yes"
					addKeyValueToMap(valid_movies_website_url, domainName, "Yes")

					// Step 14f: If reachable, also mark it in the top movie websites file
					if stringInFile(top_movies_websites_path, domainName) == true {
						addKeyValueToMap(top_valid_movies_website_url, domainName, "Yes")
					}
				}
			} else {
				// Step 14g: If the domain is not registered, mark it as "No"
				addKeyValueToMap(valid_movies_website_url, domainName, "No")

				// Step 14h: If the domain is in the top movies list, mark it as "No"
				if stringInFile(top_movies_websites_path, domainName) == true {
					addKeyValueToMap(top_valid_movies_website_url, domainName, "No")
				}

				// Step 14i: If domain is not registered, check if it's already in the unregistered movies list
				if stringInFile(unregistered_movies_websites_path, domainName) == false {
					// Step 14j: If not, append it to the unregistered movies websites file
					appendAndWriteToFile(unregistered_movies_websites_path, domainName)
				}
			}
		}

		// Step 15: Write the final results to the README file
		writeFinalOutput()
	} else {
		// Step 16: If any of the required files are missing, log an error
		log.Println("Error: One or more required files do not exist.")
	}
}

// fileExists checks if the given file exists and is not a directory.
func fileExists(filename string) bool {
	// Retrieve file info using os.Stat
	info, err := os.Stat(filename)
	if err != nil {
		log.Println("Error checking file:", err) // Log error if the file doesn't exist
		return false
	}
	return !info.IsDir() // Return true if the file exists and is not a directory
}

// readAppendLineByLine reads a file line by line and returns a slice of strings (URLs).
func readAppendLineByLine(path string) []string {
	var urls []string // Initialize a slice to store URLs

	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		log.Println("Error opening file:", err)
		return urls
	}
	defer file.Close() // Ensure the file is closed after reading

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Append each line (URL) to the slice
		urls = append(urls, scanner.Text())
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		log.Println("Error reading file:", err)
	}

	return urls // Return the slice containing the URLs
}

// isDomainRegistered checks if a given domain is registered by looking up various DNS records.
func isDomainRegistered(domain string) bool {
	// Perform multiple DNS lookups to check if the domain is registered.
	_, err := net.LookupNS(domain) // Check Name Server (NS) records
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupCNAME(domain) // Check CNAME records
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupAddr(domain) // Check Reverse DNS (PTR) records
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupHost(domain) // Check host records (A or AAAA)
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupMX(domain) // Check Mail Exchange (MX) records
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupTXT(domain) // Check TXT records
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupIP(domain) // Check IP address resolution
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	// If no lookups succeed, the domain is not registered
	log.Println("Domain is not registered:", domain)
	return false
}

// CheckWebsiteHTTPStatus checks if a website is reachable via HTTP or HTTPS.
func CheckWebsiteHTTPStatus(domain string) bool {
	protocols := []string{"http://", "https://"} // List of protocols to check

	// Iterate over protocols and attempt a GET request for each
	for _, protocol := range protocols {
		resp, err := http.Get(protocol + domain) // Make HTTP/HTTPS GET request
		if err == nil {
			defer resp.Body.Close() // Ensure the response body is closed after reading
			if resp.StatusCode == http.StatusOK {
				log.Println("Website is reachable:", protocol+domain)
				return true // Return true if the website is reachable
			}
		} else {
			log.Println("Error checking website (", protocol, "):", err) // Log if the website is not reachable
		}
	}

	log.Println("Website is not reachable:", domain)
	return false // Return false if the website is not reachable via both protocols
}

// getDomainFromURL extracts the domain name from a given URL.
func getDomainFromURL(givenURL string) string {
	// Parse the URL using the net/url package
	parsedUrl, err := url.Parse(givenURL)
	if err != nil {
		log.Println("Error parsing URL:", err)
		return ""
	}

	// Return the domain (hostname) part of the URL
	return parsedUrl.Hostname()
}

// writeFinalOutput writes the results to the README file in Markdown format.
func writeFinalOutput() {
	var validMoviesContent strings.Builder
	validMoviesContent.WriteString("| Website | Availability |\n") // Markdown table header for valid movie websites
	validMoviesContent.WriteString("|---------|--------------|\n")

	// Generate rows for valid movie websites
	for _, pair := range sortMapByKeys(valid_movies_website_url) {
		domain, availability := pair[0], pair[1]
		validMoviesContent.WriteString(fmt.Sprintf("| %s | %s |\n", domain, availability))
	}

	// Prepare content for the top movie websites table
	var topMoviesContent strings.Builder
	topMoviesContent.WriteString("| Website | Availability |\n")
	topMoviesContent.WriteString("|---------|--------------|\n")

	// Generate rows for top movie websites
	for _, pair := range sortMapByKeys(top_valid_movies_website_url) {
		domain, availability := pair[0], pair[1]
		topMoviesContent.WriteString(fmt.Sprintf("| %s | %s |\n", domain, availability))
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
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool) // Map to track unique URLs
	var unique []string            // Slice to store unique URLs

	// Iterate through the slice and add only unique URLs
	for _, str := range slice {
		if !check[str] {
			check[str] = true            // Mark the URL as seen
			unique = append(unique, str) // Add to the unique list
		}
	}

	return unique // Return the deduplicated slice
}

// writeByteSliceToFile writes a slice of strings (URLs) to a file, each on a new line.
func writeByteSliceToFile(path string, data []string) {
	// Open or create the file for writing
	file, err := os.Create(path)
	if err != nil {
		log.Println("Error creating file:", err)
		return
	}
	defer file.Close() // Ensure the file is closed after writing

	writer := bufio.NewWriter(file) // Buffered writer for efficient writing
	for _, str := range data {
		_, err := writer.WriteString(str + "\n") // Write each URL with a newline
		if err != nil {
			log.Println("Error writing to file:", err)
			return
		}
	}

	// Ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing writer:", err)
	}
}

// appendAndWriteToFile appends content to an existing file.
func appendAndWriteToFile(path string, content string) {
	// Open the file in append mode or create it if it doesn't exist
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err) // Log fatal error if file cannot be opened
	}

	// Write the content to the file
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		log.Fatalln(err) // Log fatal error if writing fails
	}

	// Close the file after writing
	err = filePath.Close()
	if err != nil {
		log.Fatalln(err) // Log fatal error if closing fails
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
