package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ung/api/internal/models"
)

// ScraperService handles job scraping from various sources
type ScraperService struct {
	client *http.Client
}

// NewScraperService creates a new scraper service
func NewScraperService() *ScraperService {
	return &ScraperService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ScrapeJobs scrapes jobs from all configured sources
func (s *ScraperService) ScrapeJobs(skills []string) ([]models.Job, error) {
	var allJobs []models.Job

	// Scrape HackerNews Who's Hiring
	hnJobs, err := s.scrapeHackerNews(skills)
	if err != nil {
		fmt.Printf("HackerNews scrape error: %v\n", err)
	} else {
		allJobs = append(allJobs, hnJobs...)
	}

	// Scrape RemoteOK
	remoteOKJobs, err := s.scrapeRemoteOK(skills)
	if err != nil {
		fmt.Printf("RemoteOK scrape error: %v\n", err)
	} else {
		allJobs = append(allJobs, remoteOKJobs...)
	}

	// Scrape WeWorkRemotely
	wwrJobs, err := s.scrapeWeWorkRemotely(skills)
	if err != nil {
		fmt.Printf("WeWorkRemotely scrape error: %v\n", err)
	} else {
		allJobs = append(allJobs, wwrJobs...)
	}

	// Scrape Jobicy
	jobicyJobs, err := s.scrapeJobicy(skills)
	if err != nil {
		fmt.Printf("Jobicy scrape error: %v\n", err)
	} else {
		allJobs = append(allJobs, jobicyJobs...)
	}

	// Scrape Arbeitnow
	arbeitnowJobs, err := s.scrapeArbeitnow(skills)
	if err != nil {
		fmt.Printf("Arbeitnow scrape error: %v\n", err)
	} else {
		allJobs = append(allJobs, arbeitnowJobs...)
	}

	// Scrape Djinni (Ukraine)
	djinniJobs, err := s.scrapeDjinni(skills)
	if err != nil {
		fmt.Printf("Djinni scrape error: %v\n", err)
	} else {
		allJobs = append(allJobs, djinniJobs...)
	}

	// Scrape DOU (Ukraine)
	douJobs, err := s.scrapeDOU(skills)
	if err != nil {
		fmt.Printf("DOU scrape error: %v\n", err)
	} else {
		allJobs = append(allJobs, douJobs...)
	}

	// Scrape European jobs (Netherlands, Germany)
	euroJobs, err := s.scrapeEuroJobs(skills)
	if err != nil {
		fmt.Printf("EuroJobs scrape error: %v\n", err)
	} else {
		allJobs = append(allJobs, euroJobs...)
	}

	return allJobs, nil
}

// HNItem represents a HackerNews item
type HNItem struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Time    int64  `json:"time"`
	Kids    []int  `json:"kids"`
	Deleted bool   `json:"deleted"`
	Dead    bool   `json:"dead"`
}

// scrapeHackerNews scrapes the latest "Who's Hiring" thread
func (s *ScraperService) scrapeHackerNews(skills []string) ([]models.Job, error) {
	// Search for the latest "Who is hiring" thread
	// HN API: https://hacker-news.firebaseio.com/v0/

	// First, get the user "whoishiring" submissions
	resp, err := s.client.Get("https://hacker-news.firebaseio.com/v0/user/whoishiring.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch whoishiring user: %w", err)
	}
	defer resp.Body.Close()

	var user struct {
		Submitted []int `json:"submitted"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	if len(user.Submitted) == 0 {
		return nil, fmt.Errorf("no submissions found")
	}

	// Find the most recent "Who is hiring" post
	var hiringPostID int
	for _, id := range user.Submitted[:min(10, len(user.Submitted))] {
		itemResp, err := s.client.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id))
		if err != nil {
			continue
		}

		var item HNItem
		if err := json.NewDecoder(itemResp.Body).Decode(&item); err != nil {
			itemResp.Body.Close()
			continue
		}
		itemResp.Body.Close()

		if strings.Contains(strings.ToLower(item.Title), "who is hiring") {
			hiringPostID = id
			break
		}
	}

	if hiringPostID == 0 {
		return nil, fmt.Errorf("no 'Who is hiring' post found")
	}

	// Get the hiring post to get comment IDs
	postResp, err := s.client.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", hiringPostID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hiring post: %w", err)
	}
	defer postResp.Body.Close()

	var hiringPost HNItem
	if err := json.NewDecoder(postResp.Body).Decode(&hiringPost); err != nil {
		return nil, fmt.Errorf("failed to decode hiring post: %w", err)
	}

	// Fetch comments (job postings) - limit to first 100 for performance
	var jobs []models.Job
	commentLimit := min(100, len(hiringPost.Kids))

	for _, commentID := range hiringPost.Kids[:commentLimit] {
		commentResp, err := s.client.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", commentID))
		if err != nil {
			continue
		}

		var comment HNItem
		if err := json.NewDecoder(commentResp.Body).Decode(&comment); err != nil {
			commentResp.Body.Close()
			continue
		}
		commentResp.Body.Close()

		if comment.Deleted || comment.Dead || comment.Text == "" {
			continue
		}

		// Parse the job posting
		job := s.parseHNJob(comment, skills)
		if job != nil {
			jobs = append(jobs, *job)
		}
	}

	return jobs, nil
}

// parseHNJob parses a HackerNews comment into a Job
func (s *ScraperService) parseHNJob(comment HNItem, skills []string) *models.Job {
	text := comment.Text

	// Extract company name (usually first line or in format "Company Name |")
	companyRegex := regexp.MustCompile(`^([^|<\n]+)`)
	companyMatch := companyRegex.FindStringSubmatch(text)
	company := ""
	if len(companyMatch) > 1 {
		company = strings.TrimSpace(companyMatch[1])
	}

	// Extract title/role
	titleRegex := regexp.MustCompile(`(?i)(hiring|looking for|seeking)\s+([^|<\n,]+)`)
	titleMatch := titleRegex.FindStringSubmatch(text)
	title := "Software Developer" // Default
	if len(titleMatch) > 2 {
		title = strings.TrimSpace(titleMatch[2])
	}

	// Check if remote
	remote := strings.Contains(strings.ToLower(text), "remote")

	// Extract location
	locationRegex := regexp.MustCompile(`(?i)(location|based in|office in)[:\s]+([^|<\n,]+)`)
	locationMatch := locationRegex.FindStringSubmatch(text)
	location := ""
	if len(locationMatch) > 2 {
		location = strings.TrimSpace(locationMatch[2])
	}

	// Calculate match score based on skills
	matchScore := s.calculateMatchScore(text, skills)

	// Only include if there's some match or no skills specified
	if len(skills) > 0 && matchScore == 0 {
		return nil
	}

	// Extract mentioned skills from text
	foundSkills := s.extractSkills(text)
	skillsJSON, _ := json.Marshal(foundSkills)

	return &models.Job{
		Source:      models.JobSourceHN,
		SourceID:    fmt.Sprintf("%d", comment.ID),
		SourceURL:   fmt.Sprintf("https://news.ycombinator.com/item?id=%d", comment.ID),
		Title:       title,
		Company:     company,
		Description: text,
		Skills:      string(skillsJSON),
		Remote:      remote,
		Location:    location,
		JobType:     "contract",
		MatchScore:  matchScore,
		PostedAt:    time.Unix(comment.Time, 0),
	}
}

// RemoteOKJob represents a job from RemoteOK API
type RemoteOKJob struct {
	ID           string   `json:"id"`
	Slug         string   `json:"slug"`
	Company      string   `json:"company"`
	Position     string   `json:"position"`
	Description  string   `json:"description"`
	Location     string   `json:"location"`
	Tags         []string `json:"tags"`
	URL          string   `json:"url"`
	SalaryMin    int      `json:"salary_min"`
	SalaryMax    int      `json:"salary_max"`
	Date         string   `json:"date"`
}

// WeWorkRemotelyJob represents a job from WeWorkRemotely
type WeWorkRemotelyJob struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Category    string `json:"category"`
	PublishedAt string `json:"published_at"`
}

// scrapeWeWorkRemotely scrapes jobs from WeWorkRemotely RSS feed
func (s *ScraperService) scrapeWeWorkRemotely(skills []string) ([]models.Job, error) {
	categories := []string{
		"programming",
		"devops-sysadmin",
		"product",
	}

	var allJobs []models.Job

	for _, category := range categories {
		url := fmt.Sprintf("https://weworkremotely.com/categories/%s.json", category)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "UNG Job Hunter/1.0")

		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}

		var result struct {
			Jobs []struct {
				ID          int    `json:"id"`
				Title       string `json:"title"`
				CompanyName string `json:"company_name"`
				Description string `json:"description"`
				URL         string `json:"url"`
				PublishedAt string `json:"published_at"`
			} `json:"jobs"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for _, wj := range result.Jobs {
			if wj.Title == "" {
				continue
			}

			matchScore := s.calculateMatchScore(wj.Description+" "+wj.Title, skills)
			if len(skills) > 0 && matchScore == 0 {
				continue
			}

			foundSkills := s.extractSkills(wj.Description)
			skillsJSON, _ := json.Marshal(foundSkills)

			postedAt, _ := time.Parse("2006-01-02T15:04:05Z", wj.PublishedAt)
			if postedAt.IsZero() {
				postedAt = time.Now()
			}

			job := models.Job{
				Source:      models.JobSourceWeWorkRemotely,
				SourceID:    fmt.Sprintf("%d", wj.ID),
				SourceURL:   wj.URL,
				Title:       wj.Title,
				Company:     wj.CompanyName,
				Description: wj.Description,
				Skills:      string(skillsJSON),
				Remote:      true,
				JobType:     "fulltime",
				MatchScore:  matchScore,
				PostedAt:    postedAt,
			}
			allJobs = append(allJobs, job)
		}
	}

	return allJobs, nil
}

// JobicyJob represents a job from Jobicy API
type JobicyJob struct {
	ID          int    `json:"id"`
	URL         string `json:"url"`
	JobTitle    string `json:"jobTitle"`
	CompanyName string `json:"companyName"`
	JobExcerpt  string `json:"jobExcerpt"`
	JobType     string `json:"jobType"`
	JobGeo      string `json:"jobGeo"`
	PubDate     string `json:"pubDate"`
}

// scrapeJobicy scrapes jobs from Jobicy API
func (s *ScraperService) scrapeJobicy(skills []string) ([]models.Job, error) {
	req, err := http.NewRequest("GET", "https://jobicy.com/api/v2/remote-jobs?count=50", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "UNG Job Hunter/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Jobicy: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Jobs []JobicyJob `json:"jobs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Jobicy response: %w", err)
	}

	var jobs []models.Job
	for _, jj := range result.Jobs {
		if jj.JobTitle == "" {
			continue
		}

		matchScore := s.calculateMatchScore(jj.JobExcerpt+" "+jj.JobTitle, skills)
		if len(skills) > 0 && matchScore == 0 {
			continue
		}

		foundSkills := s.extractSkills(jj.JobExcerpt)
		skillsJSON, _ := json.Marshal(foundSkills)

		postedAt, _ := time.Parse("2006-01-02 15:04:05", jj.PubDate)
		if postedAt.IsZero() {
			postedAt = time.Now()
		}

		jobType := "fulltime"
		if strings.Contains(strings.ToLower(jj.JobType), "contract") {
			jobType = "contract"
		} else if strings.Contains(strings.ToLower(jj.JobType), "part") {
			jobType = "parttime"
		}

		job := models.Job{
			Source:      models.JobSourceJobicy,
			SourceID:    fmt.Sprintf("%d", jj.ID),
			SourceURL:   jj.URL,
			Title:       jj.JobTitle,
			Company:     jj.CompanyName,
			Description: jj.JobExcerpt,
			Skills:      string(skillsJSON),
			Remote:      true,
			Location:    jj.JobGeo,
			JobType:     jobType,
			MatchScore:  matchScore,
			PostedAt:    postedAt,
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// ArbeitnowJob represents a job from Arbeitnow API
type ArbeitnowJob struct {
	Slug        string   `json:"slug"`
	CompanyName string   `json:"company_name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Remote      bool     `json:"remote"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
	JobTypes    []string `json:"job_types"`
	Location    string   `json:"location"`
	CreatedAt   int64    `json:"created_at"`
}

// scrapeArbeitnow scrapes jobs from Arbeitnow API
func (s *ScraperService) scrapeArbeitnow(skills []string) ([]models.Job, error) {
	req, err := http.NewRequest("GET", "https://www.arbeitnow.com/api/job-board-api", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "UNG Job Hunter/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Arbeitnow: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []ArbeitnowJob `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Arbeitnow response: %w", err)
	}

	var jobs []models.Job
	for _, aj := range result.Data {
		if aj.Title == "" {
			continue
		}

		// Prefer remote jobs
		if !aj.Remote {
			continue
		}

		matchScore := s.calculateMatchScore(aj.Description+" "+strings.Join(aj.Tags, " "), skills)
		if len(skills) > 0 && matchScore == 0 {
			continue
		}

		tagsJSON, _ := json.Marshal(aj.Tags)

		postedAt := time.Unix(aj.CreatedAt, 0)

		jobType := "fulltime"
		for _, jt := range aj.JobTypes {
			if strings.Contains(strings.ToLower(jt), "contract") {
				jobType = "contract"
				break
			} else if strings.Contains(strings.ToLower(jt), "part") {
				jobType = "parttime"
				break
			}
		}

		job := models.Job{
			Source:      models.JobSourceArbeitnow,
			SourceID:    aj.Slug,
			SourceURL:   aj.URL,
			Title:       aj.Title,
			Company:     aj.CompanyName,
			Description: aj.Description,
			Skills:      string(tagsJSON),
			Remote:      aj.Remote,
			Location:    aj.Location,
			JobType:     jobType,
			MatchScore:  matchScore,
			PostedAt:    postedAt,
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// scrapeRemoteOK scrapes jobs from RemoteOK API
func (s *ScraperService) scrapeRemoteOK(skills []string) ([]models.Job, error) {
	req, err := http.NewRequest("GET", "https://remoteok.com/api", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "UNG Job Hunter/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RemoteOK: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var remoteJobs []RemoteOKJob
	if err := json.Unmarshal(body, &remoteJobs); err != nil {
		return nil, fmt.Errorf("failed to decode RemoteOK response: %w", err)
	}

	var jobs []models.Job
	for _, rj := range remoteJobs {
		if rj.Position == "" {
			continue // Skip the first item which is usually metadata
		}

		// Calculate match score
		matchScore := s.calculateMatchScore(rj.Description+" "+strings.Join(rj.Tags, " "), skills)

		// Only include if there's some match or no skills specified
		if len(skills) > 0 && matchScore == 0 {
			continue
		}

		tagsJSON, _ := json.Marshal(rj.Tags)

		postedAt, _ := time.Parse("2006-01-02T15:04:05", rj.Date)

		job := models.Job{
			Source:      models.JobSourceRemoteOK,
			SourceID:    rj.ID,
			SourceURL:   rj.URL,
			Title:       rj.Position,
			Company:     rj.Company,
			Description: rj.Description,
			Skills:      string(tagsJSON),
			RateMin:     float64(rj.SalaryMin),
			RateMax:     float64(rj.SalaryMax),
			RateType:    "yearly",
			Currency:    "USD",
			Remote:      true,
			Location:    rj.Location,
			JobType:     "fulltime",
			MatchScore:  matchScore,
			PostedAt:    postedAt,
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// DjinniJob represents a job from Djinni.co (Ukrainian IT jobs)
type DjinniJob struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Location    string   `json:"location"`
	Experience  string   `json:"experience"`
	SalaryFrom  int      `json:"salary_from"`
	SalaryTo    int      `json:"salary_to"`
	Remote      bool     `json:"remote"`
	Tags        []string `json:"tags"`
	PublishedAt string   `json:"published_at"`
}

// scrapeDjinni scrapes IT jobs from Djinni.co (Ukraine)
func (s *ScraperService) scrapeDjinni(skills []string) ([]models.Job, error) {
	// Djinni has a public jobs listing - we'll scrape their JSON API
	// Categories: all, python, javascript, java, php, ruby, go, rust, etc.
	categories := []string{"all"}
	if len(skills) > 0 {
		// Map skills to Djinni categories
		skillMap := map[string]string{
			"go": "go", "golang": "go", "python": "python", "javascript": "javascript",
			"typescript": "javascript", "java": "java", "php": "php", "ruby": "ruby",
			"rust": "rust", "swift": "ios", "ios": "ios", "android": "android",
			"react": "javascript", "vue": "javascript", "angular": "javascript",
			"node": "javascript", "nodejs": "javascript",
		}
		for _, skill := range skills {
			if cat, ok := skillMap[strings.ToLower(skill)]; ok {
				categories = append(categories, cat)
			}
		}
	}

	var allJobs []models.Job
	seenIDs := make(map[string]bool)

	for _, category := range categories {
		url := fmt.Sprintf("https://djinni.co/jobs/?primary_keyword=%s", category)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "UNG Job Hunter/1.0")
		req.Header.Set("Accept", "text/html,application/xhtml+xml")

		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		// Parse jobs from HTML (simplified extraction)
		jobs := s.parseDjinniHTML(string(body), skills)
		for _, job := range jobs {
			if !seenIDs[job.SourceID] {
				seenIDs[job.SourceID] = true
				allJobs = append(allJobs, job)
			}
		}

		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return allJobs, nil
}

// parseDjinniHTML parses job listings from Djinni HTML
func (s *ScraperService) parseDjinniHTML(html string, skills []string) []models.Job {
	var jobs []models.Job

	// Simple regex-based extraction for job cards
	// Looking for patterns like: <a class="profile" href="/jobs/..."
	// and job titles, companies, etc.

	// Extract job URLs and basic info using string parsing
	lines := strings.Split(html, "\n")
	var currentJob *models.Job

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for job links
		if strings.Contains(line, `href="/jobs/`) && strings.Contains(line, `class="`) {
			// Extract job ID from URL
			start := strings.Index(line, `href="/jobs/`)
			if start != -1 {
				end := strings.Index(line[start+12:], `"`)
				if end != -1 {
					jobID := line[start+12 : start+12+end]
					if currentJob != nil && currentJob.Title != "" {
						jobs = append(jobs, *currentJob)
					}
					currentJob = &models.Job{
						Source:    models.JobSourceDjinni,
						SourceID:  jobID,
						SourceURL: "https://djinni.co/jobs/" + jobID,
						Remote:    true,
						Location:  "Ukraine",
						Currency:  "USD",
						PostedAt:  time.Now(),
					}
				}
			}
		}

		// Look for job title
		if currentJob != nil && currentJob.Title == "" {
			if strings.Contains(line, `class="job-list-item__title"`) || strings.Contains(line, `<h3`) {
				// Try to extract title text
				titleStart := strings.Index(line, ">")
				titleEnd := strings.LastIndex(line, "<")
				if titleStart != -1 && titleEnd > titleStart {
					title := strings.TrimSpace(line[titleStart+1 : titleEnd])
					title = strings.ReplaceAll(title, "&amp;", "&")
					if len(title) > 0 && len(title) < 200 {
						currentJob.Title = title
					}
				}
			}
		}

		// Look for company name
		if currentJob != nil && currentJob.Company == "" && strings.Contains(line, `class="company"`) {
			compStart := strings.Index(line, ">")
			compEnd := strings.LastIndex(line, "<")
			if compStart != -1 && compEnd > compStart {
				company := strings.TrimSpace(line[compStart+1 : compEnd])
				if len(company) > 0 && len(company) < 100 {
					currentJob.Company = company
				}
			}
		}

		// Look for salary info
		if currentJob != nil && strings.Contains(line, "$") {
			// Try to extract salary range like "$3000-5000"
			if matches := extractSalaryRange(line); matches != nil {
				currentJob.RateMin = matches[0]
				currentJob.RateMax = matches[1]
				currentJob.RateType = "monthly"
			}
		}
	}

	// Add last job if exists
	if currentJob != nil && currentJob.Title != "" {
		jobs = append(jobs, *currentJob)
	}

	// Calculate match scores
	for i := range jobs {
		jobs[i].MatchScore = s.calculateMatchScore(jobs[i].Title+" "+jobs[i].Description, skills)
	}

	return jobs
}

// extractSalaryRange extracts salary numbers from text like "$3000-5000" or "$3000 - $5000"
func extractSalaryRange(text string) []float64 {
	// Simple extraction - look for patterns like $XXXX
	var numbers []float64
	parts := strings.Fields(text)
	for _, part := range parts {
		part = strings.ReplaceAll(part, "$", "")
		part = strings.ReplaceAll(part, ",", "")
		part = strings.ReplaceAll(part, "-", " ")
		for _, num := range strings.Fields(part) {
			if n, err := strconv.ParseFloat(num, 64); err == nil && n > 100 && n < 1000000 {
				numbers = append(numbers, n)
			}
		}
	}
	if len(numbers) >= 2 {
		return []float64{numbers[0], numbers[1]}
	} else if len(numbers) == 1 {
		return []float64{numbers[0], numbers[0]}
	}
	return nil
}

// DOUJob represents a job from DOU.ua (Ukrainian IT community)
type DOUJob struct {
	Title       string
	Company     string
	Description string
	URL         string
	Location    string
	Salary      string
	PostedAt    time.Time
}

// scrapeDOU scrapes IT jobs from DOU.ua (Ukraine)
func (s *ScraperService) scrapeDOU(skills []string) ([]models.Job, error) {
	// DOU.ua is a popular Ukrainian IT community with job listings
	url := "https://jobs.dou.ua/vacancies/?category=Programming"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept-Language", "uk-UA,uk;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DOU: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return s.parseDOUHTML(string(body), skills), nil
}

// parseDOUHTML parses job listings from DOU.ua HTML
func (s *ScraperService) parseDOUHTML(html string, skills []string) []models.Job {
	var jobs []models.Job

	// DOU uses specific CSS classes for job listings
	// Looking for vacancy items with title, company, and details

	lines := strings.Split(html, "\n")
	var currentJob *models.Job

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for vacancy links
		if strings.Contains(line, `class="vt"`) || strings.Contains(line, `class="vacancy"`) {
			if strings.Contains(line, `href="`) {
				start := strings.Index(line, `href="`)
				if start != -1 {
					end := strings.Index(line[start+6:], `"`)
					if end != -1 {
						jobURL := line[start+6 : start+6+end]
						if strings.Contains(jobURL, "jobs.dou.ua") || strings.HasPrefix(jobURL, "/") {
							if !strings.HasPrefix(jobURL, "http") {
								jobURL = "https://jobs.dou.ua" + jobURL
							}
							if currentJob != nil && currentJob.Title != "" {
								jobs = append(jobs, *currentJob)
							}
							currentJob = &models.Job{
								Source:    models.JobSourceDOU,
								SourceID:  jobURL,
								SourceURL: jobURL,
								Remote:    false,
								Location:  "Ukraine",
								Currency:  "USD",
								PostedAt:  time.Now(),
							}
						}
					}
				}
			}
		}

		// Extract title
		if currentJob != nil && currentJob.Title == "" {
			if strings.Contains(line, `class="vt"`) {
				titleStart := strings.LastIndex(line, ">")
				titleEnd := strings.Index(line[titleStart:], "<")
				if titleStart != -1 && titleEnd > 0 {
					title := strings.TrimSpace(line[titleStart+1 : titleStart+titleEnd])
					if len(title) > 3 && len(title) < 200 {
						currentJob.Title = title
					}
				}
			}
		}

		// Extract company
		if currentJob != nil && currentJob.Company == "" {
			if strings.Contains(line, `class="company"`) {
				compStart := strings.LastIndex(line, ">")
				compEnd := strings.Index(line[compStart:], "<")
				if compStart != -1 && compEnd > 0 {
					company := strings.TrimSpace(line[compStart+1 : compStart+compEnd])
					if len(company) > 1 && len(company) < 100 {
						currentJob.Company = company
					}
				}
			}
		}

		// Extract location
		if currentJob != nil && strings.Contains(line, `class="cities"`) {
			locStart := strings.LastIndex(line, ">")
			locEnd := strings.Index(line[locStart:], "<")
			if locStart != -1 && locEnd > 0 {
				location := strings.TrimSpace(line[locStart+1 : locStart+locEnd])
				if len(location) > 0 {
					currentJob.Location = location + ", Ukraine"
				}
			}
		}

		// Check for remote indicator
		if currentJob != nil && (strings.Contains(strings.ToLower(line), "remote") || strings.Contains(strings.ToLower(line), "віддалено")) {
			currentJob.Remote = true
		}

		// Extract salary
		if currentJob != nil && strings.Contains(line, "$") {
			if matches := extractSalaryRange(line); matches != nil {
				currentJob.RateMin = matches[0]
				currentJob.RateMax = matches[1]
				currentJob.RateType = "monthly"
			}
		}
	}

	// Add last job
	if currentJob != nil && currentJob.Title != "" {
		jobs = append(jobs, *currentJob)
	}

	// Calculate match scores
	for i := range jobs {
		jobs[i].MatchScore = s.calculateMatchScore(jobs[i].Title+" "+jobs[i].Description, skills)
	}

	return jobs
}

// EuroJobsJob represents a job from European job boards
type EuroJobsJob struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Location    string   `json:"location"`
	Country     string   `json:"country"`
	Remote      bool     `json:"remote"`
	Tags        []string `json:"tags"`
}

// scrapeEuroJobs scrapes jobs from European job boards (Netherlands focus)
func (s *ScraperService) scrapeEuroJobs(skills []string) ([]models.Job, error) {
	// Use multiple European sources
	var allJobs []models.Job

	// 1. ICTergezocht.nl - Dutch IT jobs (has RSS)
	ictJobs, _ := s.scrapeICTergezocht(skills)
	allJobs = append(allJobs, ictJobs...)

	// 2. Honeypot.io - European tech jobs
	honeypotJobs, _ := s.scrapeHoneypot(skills)
	allJobs = append(allJobs, honeypotJobs...)

	return allJobs, nil
}

// scrapeICTergezocht scrapes IT jobs from ICTergezocht.nl (Netherlands)
func (s *ScraperService) scrapeICTergezocht(skills []string) ([]models.Job, error) {
	url := "https://www.ictergezocht.nl/vacatures"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "UNG Job Hunter/1.0")
	req.Header.Set("Accept-Language", "nl-NL,nl;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return s.parseICTergezochtHTML(string(body), skills), nil
}

// parseICTergezochtHTML parses jobs from ICTergezocht.nl
func (s *ScraperService) parseICTergezochtHTML(html string, skills []string) []models.Job {
	var jobs []models.Job

	lines := strings.Split(html, "\n")
	var currentJob *models.Job

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for job links
		if strings.Contains(line, `href="/vacature/`) {
			start := strings.Index(line, `href="/vacature/`)
			if start != -1 {
				end := strings.Index(line[start+16:], `"`)
				if end != -1 {
					jobID := line[start+16 : start+16+end]
					if currentJob != nil && currentJob.Title != "" {
						jobs = append(jobs, *currentJob)
					}
					currentJob = &models.Job{
						Source:    models.JobSourceNetherlands,
						SourceID:  jobID,
						SourceURL: "https://www.ictergezocht.nl/vacature/" + jobID,
						Location:  "Netherlands",
						Currency:  "EUR",
						PostedAt:  time.Now(),
					}
				}
			}
		}

		// Extract title and other fields
		if currentJob != nil {
			if currentJob.Title == "" && (strings.Contains(line, `<h2`) || strings.Contains(line, `<h3`) || strings.Contains(line, `class="title"`)) {
				titleStart := strings.LastIndex(line, ">")
				titleEnd := strings.Index(line[titleStart:], "<")
				if titleStart != -1 && titleEnd > 0 {
					title := strings.TrimSpace(line[titleStart+1 : titleStart+titleEnd])
					if len(title) > 3 && len(title) < 200 {
						currentJob.Title = title
					}
				}
			}

			if strings.Contains(strings.ToLower(line), "remote") || strings.Contains(strings.ToLower(line), "thuiswerken") {
				currentJob.Remote = true
			}
		}
	}

	if currentJob != nil && currentJob.Title != "" {
		jobs = append(jobs, *currentJob)
	}

	for i := range jobs {
		jobs[i].MatchScore = s.calculateMatchScore(jobs[i].Title, skills)
	}

	return jobs
}

// scrapeHoneypot scrapes tech jobs from Honeypot (European tech jobs)
func (s *ScraperService) scrapeHoneypot(skills []string) ([]models.Job, error) {
	// Honeypot focuses on Netherlands and Germany
	locations := []string{"netherlands", "germany"}
	var allJobs []models.Job

	for _, location := range locations {
		url := fmt.Sprintf("https://www.honeypot.io/pages/jobs/%s", location)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "UNG Job Hunter/1.0")

		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		jobs := s.parseHoneypotHTML(string(body), location, skills)
		allJobs = append(allJobs, jobs...)

		time.Sleep(300 * time.Millisecond)
	}

	return allJobs, nil
}

// parseHoneypotHTML parses jobs from Honeypot
func (s *ScraperService) parseHoneypotHTML(html string, country string, skills []string) []models.Job {
	var jobs []models.Job

	lines := strings.Split(html, "\n")
	var currentJob *models.Job

	countryName := "Netherlands"
	source := models.JobSourceNetherlands
	if country == "germany" {
		countryName = "Germany"
		source = models.JobSourceEuroJobs
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for job links
		if strings.Contains(line, `href="`) && strings.Contains(line, "job") {
			start := strings.Index(line, `href="`)
			if start != -1 {
				end := strings.Index(line[start+6:], `"`)
				if end != -1 {
					jobURL := line[start+6 : start+6+end]
					if strings.Contains(jobURL, "honeypot") || strings.HasPrefix(jobURL, "/") {
						if !strings.HasPrefix(jobURL, "http") {
							jobURL = "https://www.honeypot.io" + jobURL
						}
						if currentJob != nil && currentJob.Title != "" {
							jobs = append(jobs, *currentJob)
						}
						currentJob = &models.Job{
							Source:    source,
							SourceID:  jobURL,
							SourceURL: jobURL,
							Location:  countryName,
							Currency:  "EUR",
							Remote:    false,
							PostedAt:  time.Now(),
						}
					}
				}
			}
		}

		if currentJob != nil {
			// Extract title
			if currentJob.Title == "" && (strings.Contains(line, `<h2`) || strings.Contains(line, `<h3`)) {
				titleStart := strings.LastIndex(line, ">")
				titleEnd := strings.Index(line[titleStart:], "<")
				if titleStart != -1 && titleEnd > 0 {
					title := strings.TrimSpace(line[titleStart+1 : titleStart+titleEnd])
					if len(title) > 3 && len(title) < 200 {
						currentJob.Title = title
					}
				}
			}

			// Check for remote
			if strings.Contains(strings.ToLower(line), "remote") {
				currentJob.Remote = true
			}
		}
	}

	if currentJob != nil && currentJob.Title != "" {
		jobs = append(jobs, *currentJob)
	}

	for i := range jobs {
		jobs[i].MatchScore = s.calculateMatchScore(jobs[i].Title, skills)
	}

	return jobs
}

// calculateMatchScore calculates how well a job matches the user's skills
func (s *ScraperService) calculateMatchScore(text string, skills []string) float64 {
	if len(skills) == 0 {
		return 50.0 // Default score when no skills specified
	}

	text = strings.ToLower(text)
	matched := 0

	for _, skill := range skills {
		if strings.Contains(text, strings.ToLower(skill)) {
			matched++
		}
	}

	return (float64(matched) / float64(len(skills))) * 100
}

// extractSkills extracts common tech skills from text
func (s *ScraperService) extractSkills(text string) []string {
	commonSkills := []string{
		"go", "golang", "python", "javascript", "typescript", "react", "vue", "angular",
		"node", "nodejs", "java", "kotlin", "swift", "ios", "android", "rust", "c++",
		"ruby", "rails", "php", "laravel", "django", "flask", "spring", "docker",
		"kubernetes", "aws", "gcp", "azure", "postgresql", "mysql", "mongodb",
		"redis", "graphql", "rest", "api", "microservices", "devops", "ci/cd",
	}

	text = strings.ToLower(text)
	var found []string

	for _, skill := range commonSkills {
		if strings.Contains(text, skill) {
			found = append(found, skill)
		}
	}

	return found
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
