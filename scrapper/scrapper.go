package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// "github.com/!puerkito!bio/goquery"

type extractedJob struct {
	id string
	title string
	location string
	salary string
	summary string
}


func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/jobs?q=" + term + "&limit=50"
	start := time.Now()
	var jobs []extractedJob
	totalPages := getPages(baseURL)
	// fmt.Println(totalPages)

	c := make(chan []extractedJob)
	
	for i := 0; i < totalPages; i++ {
		// extractedJobs := getPage(i)
		go getPage(baseURL, i, c)
	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		// jobs = append(jobs, extractedJobs...)
		appendJobs(&jobs, extractedJobs)

	}

	// fmt.Println(len(jobs))
	writeJobs(jobs)
	fmt.Println("done extracting", len(jobs), "jobs in", time.Since(start))
}

func appendJobs(jobs *[]extractedJob, extractedJobs []extractedJob) {
	// fmt.Println(jobs)
	// fmt.Println(*jobs)
	*jobs = append(*jobs, extractedJobs...)
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "TITLE", "LOCATION", "SALARY", "SUMMARY"}
	wErr := w.Write(headers)
	checkErr(wErr)

	// c := make(chan error)

	for _, job := range(jobs) {
		jobSlice := []string{"kr.indeed.com/viewjob?jk="+job.id, job.title, job.location, job.salary, job.summary}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
		// go writeJob(w, job, c)
	}

	// for i := 0; i < len(jobs); i++ {
	// 	jwErr := <- c
	// 	checkErr(jwErr)
	// }
}

// func writeJob(w *csv.Writer, job extractedJob, channel chan<- error) {
// 	jobSlice := []string{"kr.indeed.com/viewjob?jk="+job.id, job.title, job.location, job.salary, job.summary}
// 	jwErr := w.Write(jobSlice)
// 	channel <- jwErr
// }

// cleans a string
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ") 
}

func getPage(baseURL string, page int, channel chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)
	fmt.Println("requesting", pageURL)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, card *goquery.Selection) {
		// job := extractJob(card)
		go extractJob(card, c)
		// jobs = append(jobs, job)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <- c 
		jobs = append(jobs, job)
	}

	// return jobs
	channel <- jobs
}

func extractJob(card *goquery.Selection, channel chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	location := CleanString(card.Find(".sjcl").Text())
	// fmt.Println(id, exists, title, location)
	salary := CleanString(card.Find(".salaryText").Text())
	summary := CleanString(card.Find(".summary").Text())

	// return extractedJob{
	// 	id: id,
	// 	title: title,
	// 	location: location,
	// 	salary: salary,
	// 	summary: summary,
	// }
	channel <- extractedJob{
		id: id,
		title: title,
		location: location,
		salary: salary,
		summary: summary,
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatusCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}

func getPages(baseURL string) int {
	pages := 0
	res, err := http.Get(baseURL)
	// if err != nil {
	// 	log.Fatalln(err)
	// } else if res.StatusCode != 200 {
	// 	log.Fatalln("Request failed with Status:", res.StatusCode)
	// }
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	// fmt.Println(doc)
	doc.Find(".pagination").Each(func(i int, sel *goquery.Selection) {
		pages = sel.Find("a").Length()
	})
	
	return pages
}