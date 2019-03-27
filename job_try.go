// job_try.go
package main

import (
	"os"
	"fmt"
	"strings"
	"bufio"
	"strconv"
	"log"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

/* Example URL's to base making URL's off of 
https://www.indeed.com/jobs?q=software+engineer&l=Denver%2C+CO

https://www.indeed.com/jobs?q=software+engineer+$70,000&l=Denver,+CO&radius=10&explvl=entry_level

https://www.indeed.com/jobs?q=software+engineer+$75,000&l=San+Diego,+CA&radius=10&jt=fulltime&explvl=entry_level
*/


// Global variable to keep track of keywords
var keywords string

// Function to make the URL for indeed web scraping
func makeIndeedURL(title string, salary string, city string, state string, radius string, jobType string, expr string, start int) string {
	// Begining of the indeed job search URL
	baseURL := "https://www.indeed.com/jobs?q="
	
	// Add title to URL
	baseURL += strings.ReplaceAll(title, " ", "+")

	// Add salary to URL
	if salary != "" {
		baseURL += "+" + salary
	}

	// Add location to URL
	baseURL += "&l=" + strings.ReplaceAll(strings.Title(strings.ToLower(city)), " ", "+")
	baseURL += ",+" + strings.ToUpper(state)
	
	// Add radius
	if radius != "" {
		baseURL += "&radius=" + radius
	}

	// Add job type
	if jobType != "" {
		baseURL += "&jt=" + strings.ReplaceAll(jobType, " ", "")
	}

	// Add experience
	if expr != "" {
		baseURL += "&explvl=" + strings.ReplaceAll(expr, " ", "_")
	}

	// Add start value to URL
	baseURL += "&start=" + strconv.Itoa(start)

	return baseURL
}

// Function to get the user input for different variables for the search
func getUserInput() (string, string, string, string, string, string, string) {
	// Create a bufio Reader to get user input
	reader := bufio.NewReader(os.Stdin)

	var jobTitle, salary, city, state, radius, jobType, experience string
	
	// Get Job title for job search
	for {
		fmt.Println("Enter the desired job title (required):")
		jobTitle, _ = reader.ReadString('\n')
		// Get rid of newline character at the end
		jobTitle = strings.TrimRight(jobTitle, "\n")
		// Loop until something is entered in the job title
		if jobTitle != "" {
			break
		} else {
			fmt.Println("Job title required.")
		}
	}

	// Get desired salary
	fmt.Println("Enter desired salary (optional, ex: $75,000):")
	salary, _ = reader.ReadString('\n')
	salary = strings.TrimRight(salary, "\n")

	// Get city and state
	for {
		fmt.Println("Enter desired city (required):")
		city, _ = reader.ReadString('\n')
		city = strings.TrimRight(city, "\n")

		fmt.Println("Enter desired state (required, ex: CO):")
		state, _ = reader.ReadString('\n')
		state = strings.TrimRight(state, "\n")
		if city != "" && state != "" {
			break
		} else {
			fmt.Println("City and State required.")
		}	
	}

	// Get Radius
	fmt.Println("Enter radius in miles (optional, default = 10, ex: 25):")
	radius, _ = reader.ReadString('\n')
	radius = strings.TrimRight(radius, "\n")
	// Set default radius to 10 miles	
	if radius == "" {
		radius = "10"
	}

	// Get job type
	for {
		fmt.Println("Enter job type (optional, options: full time, internship, part time):")
		jobType, _ = reader.ReadString('\n')
		jobType = strings.TrimRight(jobType, "\n")
		if jobType == "" || jobType == "full time" || jobType == "full time" || jobType == "internship" || jobType == "part time" || jobType == "parttime" {
			break
		} else {
			fmt.Println("Job type must be: full time, internship, part time")
		}
	}

	// Get experience level
	for {
		fmt.Println("Enter experience level (optional, options: entry level, mid level, senior level):")
		experience, _ = reader.ReadString('\n')
		experience = strings.TrimRight(experience, "\n")
		if experience == "" || experience == "entry level" || experience == "mid level" || experience == "senior level" {
			break
		} else {
			fmt.Println("Experience must be: entry level, mid level, senior level")
		}
	}

	// Get keywords for job description
	fmt.Println("Enter keywords to search description for, separated by a comma:")
	keywords, _ = reader.ReadString('\n')

	return jobTitle, salary, city, state, radius, jobType, experience
}

//https://www.indeed.com/viewjob?jk=a83ef571940ac1dd&from=serp&vjs=3

// Function to get information from the document for Indeed
func getDocInfoIndeed(idx int, element *goquery.Selection) {
	// Get the job title
	jobTitle := element.Find(".jobtitle").Text()
	if jobTitle != "" {
		fmt.Println(strings.TrimSpace(jobTitle))	
	}
	// Get the Company Name
	company := element.Find(".company").Text()
	if company != "" {
		fmt.Println(strings.TrimSpace(company))
	}	
	// Get the Company Location
	location := element.Find(".location").Text()
	if location != "" {
		fmt.Println(strings.TrimSpace(location))
	}
	// Get the salary information
	salary := element.Find(".salarySnippet").Text()
	if salary != "" {
		fmt.Println(strings.TrimSpace(salary))
	}
	// Get Job link
	jobDescrURL, hasDescrURL := element.Find(".jobtitle").Find("a").Attr("href")
	if hasDescrURL {
		jobDescrURL = "https://www.indeed.com" + jobDescrURL
		fmt.Println(jobDescrURL)
		searchJobDescriptionIndeed(jobDescrURL)
	} else {
		fmt.Println("No job description link, cannot look for keywords.")
	}

	fmt.Println()
}

// Function to search job description for certain keywords
func searchJobDescriptionIndeed(jobDescrURL string) {
	//fmt.Println("JOBLINK:", jobDescrURL)
	//fmt.Println("KEYWORDS:", keywords)

	// Get HTML for Indeed Job Description
	jobDescr, err := http.Get(jobDescrURL)
	if err != nil {
		log.Fatal(err)
	}
	defer jobDescr.Body.Close()

	// Create goquery document for job description
	jobDescrDoc, err := goquery.NewDocumentFromReader(jobDescr.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Split keywords by comma
	keywordSplits := strings.Split(keywords, ", ")

	// Search description using goquery
	jobDescrText := jobDescrDoc.Find(".jobsearch-JobComponent-description").Text()
	
	// Put the job description in lowercase, and loop through each
	// word and check if the description contains the word
	keywordCount := 0
	jobDescrText = strings.ToLower(jobDescrText)
	for _, word := range keywordSplits {
		if strings.Contains(jobDescrText, strings.ToLower(word)) {
			fmt.Println("Contains Keyword:", word)
			keywordCount += 1
		}
	}

	// Print amount of keywords found
	fmt.Println(keywordCount, "out of", len(keywordSplits), "keywords found")
}

// Main function
func main() {
	// Declare variables for user input
	jobTitle, salary, city, state, radius, jobType, experience := getUserInput()

	// To go the the next page in an indeed search page, increase
	// the start by 10
	for start := 0; start <= 50; start += 10 {
		// Get the indeed URL to search for jobs
		indeedURL := makeIndeedURL(jobTitle, salary, city, state, radius, jobType, experience, start)
	
		// Print Indeed URL
		fmt.Println("\nIndeed URL:")
		fmt.Println(indeedURL, '\n')

		// HTTP Get request for URL
		indeedResp, err := http.Get(indeedURL)
		// Check for an error getting the URL
		if err != nil {
			log.Fatal(err)
		}
		defer indeedResp.Body.Close()

		// Create a goquery document
		indeedDoc, err := goquery.NewDocumentFromReader(indeedResp.Body)
		// Check for goquery error
		if err != nil {
			log.Fatal(err)
		}

		// Find elements in the document
		indeedDoc.Find(".jobsearch-SerpJobCard").Each(getDocInfoIndeed)
	}
}


