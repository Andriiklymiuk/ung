package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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
