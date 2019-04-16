This program is a webscraper built in Go that will search job sites, and return the results based on certain user inputs and keywords.

Two libraries need to be installed first using go get:

> $ go get github.com/PuerkitoBio/goquery
> $ go get github.com/gansidui/priority_queue

To run the program from the command line:

> $ go run job_scraper.go

Once running, it will ask for different parameters from the command line for the job search. These include: job title, location, salary, job type, and experience level. It will then ask you for a list of comma separated keywords to search the job description with. These should be entered like: java, python, c++, linux. It will also ask the minimum number of keywords that need to be found in the job description in order to be returned.

After the parameters have been entered, it will begin to search different job sites based on the input parameters. It will give a progress update denoting how many jobs it has searched out of how many jobs it has found. All of the matches are stored in a priority queue based on the number of keyword matches in the description, which are then displayed in output.txt with the highest matching jobs at the top based on the keywords. If the same company has multiple job postings, it will only return the job posting with the highest number of keyword matches for that company.

Each job match will contain the title of the job, the company posting the job, the location, a link to the job posting, the salary, how many keywords were matched, as well as which keywords were found in the description.
