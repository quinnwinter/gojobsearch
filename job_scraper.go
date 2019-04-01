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
	"github.com/gansidui/priority_queue"
	"github.com/PuerkitoBio/goquery"
)

/* Example URL's to base making URL's off of 
https://www.indeed.com/jobs?q=software+engineer&l=Denver%2C+CO

https://www.indeed.com/jobs?q=software+engineer+$70,000&l=Denver,+CO&radius=10&explvl=entry_level

https://www.indeed.com/jobs?q=software+engineer+$75,000&l=San+Diego,+CA&radius=10&jt=fulltime&explvl=entry_level
*/

// Struct for keeping track of job listings
type JobListing struct {
	Company string
	Title string
	Location string
	Salary string
	JobLink string
	Description string
	Keywords []string
	KeywordMatches int
}

// Priority Queue library used from https://github.com/gansidui/priority_queue
func (this *JobListing) Less(other interface{}) bool {
	return this.KeywordMatches > other.(*JobListing).KeywordMatches
}

// Make a global priority queue
var pq = priority_queue.New()

// Global variable to keep track of keywords
var keywords string
var numKeywords int

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

	// Get desired salary
	fmt.Println("Enter desired salary (optional, ex: $75,000):")
	salary, _ = reader.ReadString('\n')
	salary = strings.TrimRight(salary, "\n")

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

// Function to get information from the document for Indeed
func getDocInfoIndeed(idx int, element *goquery.Selection) {
	// Get the job title
	jobTitle := element.Find(".jobtitle").Text()
	if jobTitle != "" {
		jobTitle = strings.TrimSpace(jobTitle)
		// fmt.Println(jobTitle)	
	}
	// Get the Company Name
	company := element.Find(".company").Text()
	if company != "" {
		company = strings.TrimSpace(company)
		// fmt.Println(company)
	}	
	// Get the Company Location
	location := element.Find(".location").Text()
	if location != "" {
		location = strings.TrimSpace(location)
		// fmt.Println(location)
	}
	// Get the salary information
	salary := element.Find(".salarySnippet").Text()
	if salary != "" {
		salary = strings.TrimSpace(salary)
		// fmt.Println(salary)
	}
	// Get Job link
	var jobDescrText string
	var jobMatches int
	var matches []string

	jobDescrURL, hasDescrURL := element.Find(".jobtitle").Find("a").Attr("href")
	if hasDescrURL {
		jobDescrURL = "https://www.indeed.com" + jobDescrURL
		// fmt.Println(jobDescrURL)
		jobDescrText, jobMatches, matches = searchJobDescriptionIndeed(jobDescrURL)
		// Create the JobListing struct and add to the priority queue
		jl := &JobListing{
			Company: company,
			Title: jobTitle,
			Location: location,
			Salary: salary,
			JobLink: jobDescrURL,
			Description: jobDescrText,
			Keywords: matches,
			KeywordMatches: jobMatches,
		}

		// Push onto the Priority Queue
		pq.Push(jl)
	} 

	// fmt.Println()
}

// Function to search job description for certain keywords
func searchJobDescriptionIndeed(jobDescrURL string) (string, int, []string) {
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
	var wordMatches = make([]string, 0, len(keywordSplits))
	jobDescrText = strings.ToLower(jobDescrText)
	for _, word := range keywordSplits {
		if strings.Contains(jobDescrText, strings.ToLower(word)) {
			// fmt.Println("Contains Keyword:", word)
			keywordCount += 1
			wordMatches = append(wordMatches, word)
		}
	}

	// Get total number of keywords
	if numKeywords == 0 {
		numKeywords = len(keywordSplits)
	}

	// Print amount of keywords found
	// fmt.Println(keywordCount, "out of", len(keywordSplits), "keywords found")

	return jobDescrText, keywordCount, wordMatches
}

// Function to check is company is in array
func containsCompany(arr []string, company string) bool {
	for _, comp := range arr {
		if comp == company {
			return true
		}
	}
	return false
}

// Main function
func main() {
	// Declare variables for user input
	jobTitle, salary, city, state, radius, jobType, experience := getUserInput()
	
	// Initialize the heap
	//heap.Init(&pq)
	//tempJob := &JobListing{KeywordMatches: -5, Index: 0}
	//pq = make(PriorityQueue, len(temp)

	// Variable to see how many jobs there are
	var jobCount int = 1

	// To go the the next page in an indeed search page, increase
	// the start by 10
	// TODO PUT JOB COUNT AS STOPPING CONDITION
	for start := 0; start <= 10 /*jobCount*/; start += 10 {
		// Get the indeed URL to search for jobs
		indeedURL := makeIndeedURL(jobTitle, salary, city, state, radius, jobType, experience, start)
	
		// Print Indeed URL for testing
		// fmt.Println("\nIndeed URL:")
		// fmt.Println(indeedURL, '\n')

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

		// Get number of jobs from the website
		if jobCount <= 1 {
			numJobs := indeedDoc.Find("#searchCount").Text()
			numJobs = strings.TrimSpace(numJobs)
			searchCount := strings.Split(numJobs, " ")
			maxJobs := searchCount[3]
			maxJobs = strings.ReplaceAll(maxJobs, ",", "")
			jobCount, _ = strconv.Atoi(maxJobs)
		}

		// Print status of jobs searched
		fmt.Println("Jobs searched:", start, "out of", jobCount)

		// Find elements in the document
		indeedDoc.Find(".jobsearch-SerpJobCard").Each(getDocInfoIndeed)
	}

	// Get a list of the jobs, and only return the highest match job
	var companyList = make([]string, 0, pq.Len())

	// Create a file to write to
	file, err := os.Create("job_output.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Get JobListings From Priority Queue
	for pq.Len() > 0 {
		job := pq.Pop().(*JobListing)
		if !containsCompany(companyList, job.Company) {
			companyList = append(companyList, job.Company)
			fmt.Println("Title:", job.Title)
			fmt.Println("Company:", job.Company)
			fmt.Println("Location:", job.Location)
			fmt.Println("Link:", job.JobLink)
			fmt.Println("Salary:", job.Salary)
			fmt.Println("Matches:", job.KeywordMatches, "out of", numKeywords, "keywords matched")
			fmt.Println("Keywords:", strings.Join(job.Keywords, ", "))
			fmt.Println()  
			
			// Write to file
			fmt.Fprintln(file, "Title:", job.Title)
			fmt.Fprintln(file, "Company:", job.Company)
			fmt.Fprintln(file, "Location:", job.Location)
			fmt.Fprintln(file, "Link:", job.JobLink)
			fmt.Fprintln(file, "Salary:", job.Salary)
			fmt.Fprintln(file, "Matches:", job.KeywordMatches, "out of", numKeywords, "keywords matched")
			fmt.Fprintln(file, "Keywords:", strings.Join(job.Keywords, ", "))
			fmt.Fprintln(file)
		}
	}
}


