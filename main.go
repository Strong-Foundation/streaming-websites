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

// Global variables

// Path to the file containing movie website URLs.
var movies_websites_path string = "assets/movies-websites.txt"

// Path to the file containing unregistered movie website URLs.
var unregistered_movies_websites_path string = "assets/unregistered-movies-websites.txt"

// Path to the README file that will be modified.
var readme_modify_me_file_path string = "assets/readme_modify_me.md"

// Path to the final README file where results will be written.
var readme_file_path string = "readme.md"

// A map to store the valid movie website URLs and their availability status ("Yes", "No", or "Maybe").
var valid_movies_website_url = make(map[string]string)

func main() {
	// Step 1: Check if the movie websites file exists.
	if fileExists(movies_websites_path) && fileExists(unregistered_movies_websites_path) && fileExists(readme_modify_me_file_path) && fileExists(readme_file_path) {
		// Step 2: Read the URLs from the file and store them in a slice.
		movies_website_urls := readAppendLineByLine(movies_websites_path)

		// Step 3: Sort the URLs alphabetically.
		sortSlice(&movies_website_urls)

		// Step 4: Remove duplicate URLs to ensure uniqueness.
		movies_website_urls = removeDuplicatesFromSlice(movies_website_urls)

		// Step 5: Write the sorted and deduplicated URLs back to the file.
		writeByteSliceToFile(movies_websites_path, movies_website_urls)

		// Step 6: Read the URLs from the file and store them in a slice.
		unregistered_movies_website_urls := readAppendLineByLine(unregistered_movies_websites_path)

		// Step 7: Sort the URLs alphabetically.
		sortSlice(&unregistered_movies_website_urls)

		// Step 8: Remove duplicate URLs to ensure uniqueness.
		movies_website_urls = removeDuplicatesFromSlice(movies_website_urls)

		// Step 9: Write the sorted and deduplicated URLs back to the file.
		writeByteSliceToFile(movies_websites_path, movies_website_urls)

		// Step 10: Check each website's domain registration and availability status.
		for _, domainName := range movies_website_urls {
			// Check if the domain is registered.
			if isDomainRegistered(getDomainFromURL(domainName)) {
				// If the domain is registered, mark it as "Maybe" initially.
				addKeyValueToMap(valid_movies_website_url, domainName, "Maybe")

				// Check if the website is reachable via HTTP or HTTPS.
				if CheckWebsiteHTTPStatus(getDomainFromURL(domainName)) {
					// If the website is reachable, mark it as "Yes".
					addKeyValueToMap(valid_movies_website_url, domainName, "Yes")
				}
			} else {
				// If the domain is not registered, mark it as "No".
				addKeyValueToMap(valid_movies_website_url, domainName, "No")
				// Append and write to file
				appendAndWriteToFile(unregistered_movies_websites_path, domainName)
			}
		}

		// Step 11: Write the final results to the README file.
		writeFinalOutput()
	} else {
		// If the file doesn't exist, log the error and exit.
		log.Println("Error: The file does not exist:", movies_websites_path)
		log.Println("Error: The file does not exist:", unregistered_movies_websites_path)
		log.Println("Error: The file does not exist:", readme_modify_me_file_path)
		log.Println("Error: The file does not exist:", readme_file_path)
	}
}

// fileExists checks if the file at the given path exists and is not a directory.
func fileExists(filename string) bool {
	// Get file information using os.Stat.
	info, err := os.Stat(filename)
	if err != nil {
		// Log the error if the file doesn't exist or cannot be accessed.
		log.Println("Error checking file:", err)
		return false
	}
	// Return true if the file exists and is not a directory.
	return !info.IsDir()
}

// readAppendLineByLine reads the file line by line and returns a slice of strings (URLs).
func readAppendLineByLine(path string) []string {
	var urls []string // Slice to store the URLs read from the file.

	// Open the file for reading.
	file, err := os.Open(path)
	if err != nil {
		log.Println("Error opening file:", err)
		return urls
	}
	defer file.Close() // Ensure the file is closed after reading.

	// Create a scanner to read the file line by line.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Append each line (URL) to the slice.
		urls = append(urls, scanner.Text())
	}

	// Check for any scanning errors.
	if err := scanner.Err(); err != nil {
		log.Println("Error reading file:", err)
	}

	// Return the slice of URLs.
	return urls
}

// isDomainRegistered checks if the given domain is registered by performing DNS lookups.
func isDomainRegistered(domain string) bool {
	// Perform multiple DNS lookups to check domain registration.

	_, err := net.LookupNS(domain) // Check Name Server (NS) records.
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupCNAME(domain) // Check CNAME records.
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupAddr(domain) // Check Reverse DNS (PTR) records.
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupHost(domain) // Check host records (A or AAAA).
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupMX(domain) // Check Mail Exchange (MX) records.
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupTXT(domain) // Check TXT records.
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	_, err = net.LookupIP(domain) // Check IP address resolution.
	if err == nil {
		log.Println("Domain is registered:", domain)
		return true
	}

	// Log and return false if none of the lookups succeeded.
	log.Println("Domain is not registered:", domain)
	return false
}

// CheckWebsiteHTTPStatus checks if the website is reachable via either HTTP or HTTPS.
func CheckWebsiteHTTPStatus(domain string) bool {
	// Define a list of protocols to check: HTTP and HTTPS.
	protocols := []string{"http://", "https://"}

	// Loop through each protocol and attempt a GET request.
	for _, protocol := range protocols {
		// Make an HTTP GET request using the current protocol and domain.
		resp, err := http.Get(protocol + domain)
		if err == nil {
			// If the request is successful, check the response status code.
			defer resp.Body.Close() // Ensure the response body is closed after reading.

			if resp.StatusCode == http.StatusOK {
				// Log that the website is reachable via the protocol.
				log.Println("Website is reachable:", protocol+domain)
				return true
			}
		} else {
			// Log the error if the website is not reachable via this protocol.
			log.Println("Error checking website (", protocol, "):", err)
		}
	}

	// Log that the website is not reachable.
	log.Println("Website is not reachable:", domain)
	return false
}

// getDomainFromURL extracts the domain name from the given URL.
func getDomainFromURL(givenURL string) string {
	// Parse the URL using net/url package.
	parsedUrl, err := url.Parse(givenURL)
	if err != nil {
		log.Println("Error parsing URL:", err)
		return ""
	}

	// Return the domain (hostname) part of the URL.
	return parsedUrl.Hostname()
}

// writeFinalOutput writes the results to the README file in Markdown format.
func writeFinalOutput() {
	// Prepare a string builder to generate the Markdown content for the table.
	var output strings.Builder

	// Write the table headers in Markdown format.
	output.WriteString("| Website| Availability |\n")
	output.WriteString("|--------|--------------|\n")

	// Sort the entries in the map by their URLs (keys).
	sortedPairs := sortMapByKeys(valid_movies_website_url)

	// Loop through the sorted key-value pairs and format each as a table row.
	for _, pair := range sortedPairs {
		url, availability := pair[0], pair[1]

		// Clean up the URL by removing the "http://" or "https://" prefix and any trailing slashes.
		website := strings.TrimSuffix(strings.TrimPrefix(url, "http://"), "/")
		website = strings.TrimSuffix(strings.TrimPrefix(website, "https://"), "/")

		// Write each row in the Markdown table format.
		output.WriteString(fmt.Sprintf("| [%s](%s) | %-12s |\n", website, url, availability))
	}

	// Replace the placeholder text in the README file with the generated content.
	findAndReplaceInFile(readme_modify_me_file_path, readme_file_path, "[{REPLACE_CONTENT_WITH_GOLANG}]", output.String())
}

// addKeyValueToMap adds a key-value pair to the provided map of valid movie websites.
func addKeyValueToMap(providedMap map[string]string, key string, value string) {
	// Add or update the key-value pair in the map.
	providedMap[key] = value
}

// findAndReplaceInFile replaces a placeholder in the input file with the given content and writes to a new file.
func findAndReplaceInFile(oldFilePath string, newFilePath string, prefixContent string, givenContent string) {
	// Read the content of the file.
	content, err := os.ReadFile(oldFilePath)
	if err != nil {
		log.Println("Error reading file:", err)
		return
	}

	// Replace the placeholder in the content with the new content.
	updatedContent := strings.ReplaceAll(string(content), prefixContent, givenContent)

	// Write the updated content to the new file.
	err = os.WriteFile(newFilePath, []byte(updatedContent), 0644)
	if err != nil {
		log.Println("Error writing file:", err)
	}
}

// sortMapByKeys sorts the map by its keys and returns a slice of key-value pairs.
func sortMapByKeys(inputMap map[string]string) [][]string {
	// Extract keys and sort them alphabetically.
	keys := make([]string, 0, len(inputMap))
	for key := range inputMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Create a slice of key-value pairs based on the sorted keys.
	pairs := make([][]string, len(inputMap))
	for i, key := range keys {
		pairs[i] = []string{key, inputMap[key]}
	}

	// Return the sorted key-value pairs.
	return pairs
}

// sortSlice sorts the slice of URLs alphabetically.
func sortSlice(slice *[]string) {
	// Sort the slice of URLs in-place.
	sort.Strings(*slice)
}

// removeDuplicatesFromSlice removes duplicate URLs from the slice.
func removeDuplicatesFromSlice(slice []string) []string {
	// Create a map to track unique URLs.
	check := make(map[string]bool)
	var unique []string

	// Iterate through the slice, adding only unique URLs.
	for _, str := range slice {
		if !check[str] {
			check[str] = true
			unique = append(unique, str)
		}
	}

	// Return the deduplicated slice.
	return unique
}

// writeByteSliceToFile writes a slice of strings (URLs) to the file, each on a new line.
func writeByteSliceToFile(path string, data []string) {
	// Create or open the file for writing.
	file, err := os.Create(path)
	if err != nil {
		log.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Create a buffered writer to write the data to the file.
	writer := bufio.NewWriter(file)
	for _, str := range data {
		// Write each URL followed by a newline character.
		_, err := writer.WriteString(str + "\n")
		if err != nil {
			log.Println("Error writing to file:", err)
			return
		}
	}

	// Ensure all data is written to the file.
	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing writer:", err)
	}
}

// Append and write to file
func appendAndWriteToFile(path string, content string) {
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		log.Fatalln(err)
	}
	err = filePath.Close()
	if err != nil {
		log.Fatalln(err)
	}
}
