package services

import (
	"context"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SchedulerService handles recurring tasks and notifications
type SchedulerService struct {
	db           *gorm.DB
	emailService *EmailService
	tasks        map[string]*ScheduledTask
	mu           sync.RWMutex
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// ScheduledTask represents a recurring task
type ScheduledTask struct {
	Name     string
	Interval time.Duration
	Handler  func(context.Context) error
	Enabled  bool
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(db *gorm.DB, emailService *EmailService) *SchedulerService {
	return &SchedulerService{
		db:           db,
		emailService: emailService,
		tasks:        make(map[string]*ScheduledTask),
		stopChan:     make(chan struct{}),
	}
}

// RegisterTask registers a new scheduled task
func (s *SchedulerService) RegisterTask(name string, interval time.Duration, handler func(context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[name] = &ScheduledTask{
		Name:     name,
		Interval: interval,
		Handler:  handler,
		Enabled:  true,
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start() {
	log.Println("Starting scheduler service...")

	// Register default tasks
	s.registerDefaultTasks()

	// Start all tasks
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, task := range s.tasks {
		if task.Enabled {
			s.wg.Add(1)
			go s.runTask(task)
		}
	}

	log.Printf("Scheduler started with %d tasks\n", len(s.tasks))
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	log.Println("Stopping scheduler service...")
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

// runTask runs a scheduled task at specified intervals
func (s *SchedulerService) runTask(task *ScheduledTask) {
	defer s.wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	log.Printf("Started task: %s (interval: %s)\n", task.Name, task.Interval)

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := task.Handler(ctx); err != nil {
				log.Printf("Task %s failed: %v\n", task.Name, err)
			}
		case <-s.stopChan:
			log.Printf("Stopped task: %s\n", task.Name)
			return
		}
	}
}

// registerDefaultTasks registers all default scheduled tasks
func (s *SchedulerService) registerDefaultTasks() {
	// Invoice payment reminders - runs daily at 9 AM
	s.RegisterTask("invoice_reminders", 24*time.Hour, s.sendInvoiceReminders)

	// Overdue invoice notifications - runs daily at 10 AM
	s.RegisterTask("overdue_invoices", 24*time.Hour, s.sendOverdueNotifications)

	// Contract expiry reminders - runs daily
	s.RegisterTask("contract_expiry", 24*time.Hour, s.sendContractExpiryReminders)

	// Weekly summary - runs weekly on Monday
	s.RegisterTask("weekly_summary", 7*24*time.Hour, s.sendWeeklySummary)

	// Monthly reports - runs monthly on 1st
	s.RegisterTask("monthly_reports", 30*24*time.Hour, s.sendMonthlyReports)
}

// sendInvoiceReminders sends payment reminders for upcoming due invoices
func (s *SchedulerService) sendInvoiceReminders(ctx context.Context) error {
	log.Println("Running invoice reminders task...")

	// TODO: Query invoices due in next 7 days that are pending
	// For each invoice, send reminder email to client

	/*
	Example query:
	var invoices []Invoice
	err := s.db.WithContext(ctx).
		Where("status = ? AND due_date BETWEEN ? AND ?",
			"pending",
			time.Now(),
			time.Now().AddDate(0, 0, 7)).
		Preload("Client").
		Find(&invoices).Error

	if err != nil {
		return fmt.Errorf("failed to query invoices: %w", err)
	}

	for _, invoice := range invoices {
		email := &Email{
			To:      []string{invoice.Client.Email},
			Subject: fmt.Sprintf("Payment Reminder: Invoice #%s", invoice.InvoiceNum),
			HTMLBody: s.generateInvoiceReminderHTML(invoice),
			Body:    s.generateInvoiceReminderText(invoice),
		}

		if err := s.emailService.Send(email); err != nil {
			log.Printf("Failed to send reminder for invoice %s: %v", invoice.InvoiceNum, err)
		} else {
			log.Printf("Sent reminder for invoice %s to %s", invoice.InvoiceNum, invoice.Client.Email)
		}
	}
	*/

	log.Println("Invoice reminders task completed")
	return nil
}

// sendOverdueNotifications sends notifications for overdue invoices
func (s *SchedulerService) sendOverdueNotifications(ctx context.Context) error {
	log.Println("Running overdue invoices task...")

	// TODO: Query invoices that are overdue
	// Send notification to both client and admin

	log.Println("Overdue invoices task completed")
	return nil
}

// sendContractExpiryReminders sends reminders for expiring contracts
func (s *SchedulerService) sendContractExpiryReminders(ctx context.Context) error {
	log.Println("Running contract expiry reminders task...")

	// TODO: Query contracts expiring in next 30 days
	// Send reminder to both client and admin

	log.Println("Contract expiry task completed")
	return nil
}

// sendWeeklySummary sends weekly summary to admin
func (s *SchedulerService) sendWeeklySummary(ctx context.Context) error {
	log.Println("Running weekly summary task...")

	// TODO: Generate weekly summary
	// - Invoices created this week
	// - Payments received
	// - Outstanding invoices
	// - Active time tracking sessions

	log.Println("Weekly summary task completed")
	return nil
}

// sendMonthlyReports sends monthly reports
func (s *SchedulerService) sendMonthlyReports(ctx context.Context) error {
	log.Println("Running monthly reports task...")

	// TODO: Generate monthly report
	// - Total revenue
	// - Total hours worked
	// - Client breakdown
	// - Payment status

	log.Println("Monthly reports task completed")
	return nil
}

// generateInvoiceReminderHTML generates HTML email for invoice reminder
func (s *SchedulerService) generateInvoiceReminderHTML(invoice interface{}) string {
	// TODO: Use proper template
	return `
	<!DOCTYPE html>
	<html>
	<head>
		<style>
			body { font-family: Arial, sans-serif; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
			.content { padding: 20px; background-color: #f9f9f9; }
			.button { background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; margin-top: 20px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>Payment Reminder</h1>
			</div>
			<div class="content">
				<p>Dear Client,</p>
				<p>This is a friendly reminder that your payment is due soon.</p>
				<p><strong>Invoice Details:</strong></p>
				<ul>
					<li>Invoice Number: #INV-001</li>
					<li>Amount Due: $1,000.00</li>
					<li>Due Date: 2024-01-15</li>
				</ul>
				<p>Please process your payment by the due date to avoid late fees.</p>
				<a href="#" class="button">View Invoice</a>
				<p>Thank you for your business!</p>
			</div>
		</div>
	</body>
	</html>
	`
}

// generateInvoiceReminderText generates plain text email for invoice reminder
func (s *SchedulerService) generateInvoiceReminderText(invoice interface{}) string {
	return `
Payment Reminder

Dear Client,

This is a friendly reminder that your payment is due soon.

Invoice Details:
- Invoice Number: #INV-001
- Amount Due: $1,000.00
- Due Date: 2024-01-15

Please process your payment by the due date to avoid late fees.

Thank you for your business!
	`
}
