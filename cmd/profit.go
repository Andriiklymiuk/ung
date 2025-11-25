package cmd

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var profitCmd = &cobra.Command{
	Use:   "profit",
	Short: "Interactive profit dashboard",
	Long:  "View an interactive dashboard showing profit, revenue, expenses, and trends",
	RunE:  runProfitDashboard,
}

func init() {
	rootCmd.AddCommand(profitCmd)
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)

	highlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	positiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	negativeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("229"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

type profitModel struct {
	data         *dashboardData
	selectedTab  int
	selectedRow  int
	tabs         []string
	width        int
	height       int
	loading      bool
	err          error
}

type dashboardData struct {
	// Current period
	currentRevenue  float64
	currentExpenses float64
	currentProfit   float64

	// Previous period (for comparison)
	prevRevenue  float64
	prevExpenses float64
	prevProfit   float64

	// Yearly
	yearRevenue  float64
	yearExpenses float64
	yearProfit   float64

	// Breakdown
	topClients     []clientRevenue
	expensesByType map[string]float64
	monthlyTrend   []monthData

	// Goals
	monthlyGoal    float64
	goalProgress   float64
}

type clientRevenue struct {
	name    string
	revenue float64
}

type monthData struct {
	month    string
	revenue  float64
	expenses float64
	profit   float64
}

type dataLoadedMsg struct {
	data *dashboardData
	err  error
}

func runProfitDashboard(cmd *cobra.Command, args []string) error {
	p := tea.NewProgram(initialProfitModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func initialProfitModel() profitModel {
	return profitModel{
		tabs:        []string{"Overview", "Clients", "Expenses", "Trends"},
		selectedTab: 0,
		loading:     true,
	}
}

func (m profitModel) Init() tea.Cmd {
	return loadDashboardData
}

func loadDashboardData() tea.Msg {
	data := &dashboardData{
		expensesByType: make(map[string]float64),
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfPrevMonth := startOfMonth.AddDate(0, -1, 0)
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	// Current month revenue (paid invoices)
	var currentRev sql.NullFloat64
	db.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM invoices
		WHERE status = ? AND updated_at >= ? AND updated_at < ?
	`, models.StatusPaid, startOfMonth, now).Scan(&currentRev)
	data.currentRevenue = currentRev.Float64

	// Current month expenses
	var currentExp sql.NullFloat64
	db.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM expenses
		WHERE date >= ? AND date < ?
	`, startOfMonth, now).Scan(&currentExp)
	data.currentExpenses = currentExp.Float64
	data.currentProfit = data.currentRevenue - data.currentExpenses

	// Previous month
	var prevRev sql.NullFloat64
	db.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM invoices
		WHERE status = ? AND updated_at >= ? AND updated_at < ?
	`, models.StatusPaid, startOfPrevMonth, startOfMonth).Scan(&prevRev)
	data.prevRevenue = prevRev.Float64

	var prevExp sql.NullFloat64
	db.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM expenses
		WHERE date >= ? AND date < ?
	`, startOfPrevMonth, startOfMonth).Scan(&prevExp)
	data.prevExpenses = prevExp.Float64
	data.prevProfit = data.prevRevenue - data.prevExpenses

	// Year to date
	var yearRev sql.NullFloat64
	db.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM invoices
		WHERE status = ? AND updated_at >= ?
	`, models.StatusPaid, startOfYear).Scan(&yearRev)
	data.yearRevenue = yearRev.Float64

	var yearExp sql.NullFloat64
	db.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM expenses
		WHERE date >= ?
	`, startOfYear).Scan(&yearExp)
	data.yearExpenses = yearExp.Float64
	data.yearProfit = data.yearRevenue - data.yearExpenses

	// Top clients
	rows, _ := db.DB.Raw(`
		SELECT c.name, COALESCE(SUM(i.amount), 0) as total
		FROM clients c
		LEFT JOIN invoice_recipients ir ON c.id = ir.client_id
		LEFT JOIN invoices i ON ir.invoice_id = i.id AND i.status = ?
		GROUP BY c.id, c.name
		HAVING total > 0
		ORDER BY total DESC
		LIMIT 5
	`, models.StatusPaid).Rows()
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var total float64
			rows.Scan(&name, &total)
			data.topClients = append(data.topClients, clientRevenue{name: name, revenue: total})
		}
	}

	// Expenses by category
	expRows, _ := db.DB.Raw(`
		SELECT category, SUM(amount) as total
		FROM expenses
		WHERE date >= ?
		GROUP BY category
		ORDER BY total DESC
	`, startOfYear).Rows()
	if expRows != nil {
		defer expRows.Close()
		for expRows.Next() {
			var category string
			var total float64
			expRows.Scan(&category, &total)
			data.expensesByType[category] = total
		}
	}

	// Monthly trend (last 6 months)
	for i := 5; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0)

		var rev, exp sql.NullFloat64
		db.DB.Raw(`SELECT COALESCE(SUM(amount), 0) FROM invoices WHERE status = ? AND updated_at >= ? AND updated_at < ?`,
			models.StatusPaid, monthStart, monthEnd).Scan(&rev)
		db.DB.Raw(`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE date >= ? AND date < ?`,
			monthStart, monthEnd).Scan(&exp)

		data.monthlyTrend = append(data.monthlyTrend, monthData{
			month:    monthStart.Format("Jan"),
			revenue:  rev.Float64,
			expenses: exp.Float64,
			profit:   rev.Float64 - exp.Float64,
		})
	}

	// Get monthly goal if set
	var goal IncomeGoal
	if err := db.DB.Where("period = ? AND year = ? AND month = ?",
		"monthly", now.Year(), int(now.Month())).First(&goal).Error; err == nil {
		data.monthlyGoal = goal.Amount
		data.goalProgress = (data.currentRevenue / data.monthlyGoal) * 100
		if data.goalProgress > 100 {
			data.goalProgress = 100
		}
	}

	return dataLoadedMsg{data: data}
}

func (m profitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dataLoadedMsg:
		m.loading = false
		m.data = msg.data
		m.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "right", "l":
			m.selectedTab = (m.selectedTab + 1) % len(m.tabs)
			m.selectedRow = 0
		case "shift+tab", "left", "h":
			m.selectedTab = (m.selectedTab - 1 + len(m.tabs)) % len(m.tabs)
			m.selectedRow = 0
		case "down", "j":
			m.selectedRow++
		case "up", "k":
			if m.selectedRow > 0 {
				m.selectedRow--
			}
		case "r":
			m.loading = true
			return m, loadDashboardData
		}
	}

	return m, nil
}

func (m profitModel) View() string {
	if m.loading {
		return "\n  Loading dashboard..."
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v", m.err)
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("  Profit Dashboard"))
	b.WriteString("\n\n")

	// Tabs
	var tabs []string
	for i, tab := range m.tabs {
		if i == m.selectedTab {
			tabs = append(tabs, selectedStyle.Render(" "+tab+" "))
		} else {
			tabs = append(tabs, dimStyle.Render(" "+tab+" "))
		}
	}
	b.WriteString("  " + strings.Join(tabs, "  "))
	b.WriteString("\n\n")

	// Content based on selected tab
	switch m.selectedTab {
	case 0:
		b.WriteString(m.renderOverview())
	case 1:
		b.WriteString(m.renderClients())
	case 2:
		b.WriteString(m.renderExpenses())
	case 3:
		b.WriteString(m.renderTrends())
	}

	// Help
	b.WriteString(helpStyle.Render("\n  ←/→ switch tabs • ↑/↓ navigate • r refresh • q quit"))

	return b.String()
}

func (m profitModel) renderOverview() string {
	d := m.data
	var b strings.Builder

	// Summary boxes
	b.WriteString("  ")
	b.WriteString(m.renderMetricBox("This Month", d.currentRevenue, d.currentExpenses, d.currentProfit, d.prevProfit))
	b.WriteString("  ")
	b.WriteString(m.renderMetricBox("Last Month", d.prevRevenue, d.prevExpenses, d.prevProfit, 0))
	b.WriteString("  ")
	b.WriteString(m.renderMetricBox("Year to Date", d.yearRevenue, d.yearExpenses, d.yearProfit, 0))
	b.WriteString("\n\n")

	// Goal progress
	if d.monthlyGoal > 0 {
		b.WriteString("  Goal Progress\n")
		b.WriteString("  " + m.renderProgressBar(d.goalProgress, 40))
		b.WriteString(fmt.Sprintf(" %.0f%%\n", d.goalProgress))
		b.WriteString(fmt.Sprintf("  $%.0f / $%.0f\n", d.currentRevenue, d.monthlyGoal))
	}

	// Quick stats
	b.WriteString("\n  Quick Stats\n")
	b.WriteString("  ─────────────────────────\n")

	profitMargin := 0.0
	if d.yearRevenue > 0 {
		profitMargin = (d.yearProfit / d.yearRevenue) * 100
	}

	monthChange := 0.0
	if d.prevProfit != 0 {
		monthChange = ((d.currentProfit - d.prevProfit) / absFloat(d.prevProfit)) * 100
	}

	b.WriteString(fmt.Sprintf("  Profit margin:  %.1f%%\n", profitMargin))

	if monthChange >= 0 {
		b.WriteString(fmt.Sprintf("  Month change:   %s\n", positiveStyle.Render(fmt.Sprintf("+%.1f%%", monthChange))))
	} else {
		b.WriteString(fmt.Sprintf("  Month change:   %s\n", negativeStyle.Render(fmt.Sprintf("%.1f%%", monthChange))))
	}

	avgMonthly := d.yearRevenue / float64(time.Now().Month())
	b.WriteString(fmt.Sprintf("  Avg monthly:    $%.0f\n", avgMonthly))

	return b.String()
}

func (m profitModel) renderMetricBox(title string, revenue, expenses, profit, prevProfit float64) string {
	var content strings.Builder

	content.WriteString(highlightStyle.Render(title) + "\n")
	content.WriteString(fmt.Sprintf("Revenue:  $%.0f\n", revenue))
	content.WriteString(fmt.Sprintf("Expenses: $%.0f\n", expenses))

	profitStr := fmt.Sprintf("Profit:   $%.0f", profit)
	if profit >= 0 {
		content.WriteString(positiveStyle.Render(profitStr))
	} else {
		content.WriteString(negativeStyle.Render(profitStr))
	}

	if prevProfit != 0 {
		change := profit - prevProfit
		if change >= 0 {
			content.WriteString(positiveStyle.Render(fmt.Sprintf(" (+$%.0f)", change)))
		} else {
			content.WriteString(negativeStyle.Render(fmt.Sprintf(" (-$%.0f)", -change)))
		}
	}

	return boxStyle.Width(22).Render(content.String())
}

func (m profitModel) renderClients() string {
	d := m.data
	var b strings.Builder

	b.WriteString("  Top Clients by Revenue\n")
	b.WriteString("  ───────────────────────────────────\n")

	if len(d.topClients) == 0 {
		b.WriteString("  No client revenue data\n")
		return b.String()
	}

	maxRevenue := d.topClients[0].revenue
	for i, c := range d.topClients {
		style := dimStyle
		if i == m.selectedRow {
			style = selectedStyle
		}

		bar := m.renderProgressBar((c.revenue/maxRevenue)*100, 20)
		line := fmt.Sprintf("  %-20s %s $%.0f", truncateStr(c.name, 20), bar, c.revenue)
		b.WriteString(style.Render(line) + "\n")
	}

	return b.String()
}

func (m profitModel) renderExpenses() string {
	d := m.data
	var b strings.Builder

	b.WriteString("  Expenses by Category (YTD)\n")
	b.WriteString("  ───────────────────────────────────\n")

	if len(d.expensesByType) == 0 {
		b.WriteString("  No expenses recorded\n")
		return b.String()
	}

	// Find max for scaling
	var maxExp float64
	for _, v := range d.expensesByType {
		if v > maxExp {
			maxExp = v
		}
	}

	i := 0
	for cat, amount := range d.expensesByType {
		style := dimStyle
		if i == m.selectedRow {
			style = selectedStyle
		}

		bar := m.renderProgressBar((amount/maxExp)*100, 20)
		line := fmt.Sprintf("  %-15s %s $%.0f", truncateStr(cat, 15), bar, amount)
		b.WriteString(style.Render(line) + "\n")
		i++
	}

	b.WriteString(fmt.Sprintf("\n  Total: $%.0f\n", d.yearExpenses))

	return b.String()
}

func (m profitModel) renderTrends() string {
	d := m.data
	var b strings.Builder

	b.WriteString("  6-Month Trend\n")
	b.WriteString("  ───────────────────────────────────\n\n")

	if len(d.monthlyTrend) == 0 {
		b.WriteString("  No trend data available\n")
		return b.String()
	}

	// Find max values for scaling
	var maxVal float64
	for _, md := range d.monthlyTrend {
		if md.revenue > maxVal {
			maxVal = md.revenue
		}
		if md.expenses > maxVal {
			maxVal = md.expenses
		}
	}

	// Header
	b.WriteString("       ")
	for _, md := range d.monthlyTrend {
		b.WriteString(fmt.Sprintf(" %s  ", md.month))
	}
	b.WriteString("\n")

	// Revenue bars (ASCII chart)
	b.WriteString("  Rev  ")
	for _, md := range d.monthlyTrend {
		height := int((md.revenue / maxVal) * 5)
		b.WriteString(positiveStyle.Render(fmt.Sprintf(" %s ", strings.Repeat("█", height)+strings.Repeat("░", 5-height))))
	}
	b.WriteString("\n")

	// Expense bars
	b.WriteString("  Exp  ")
	for _, md := range d.monthlyTrend {
		height := int((md.expenses / maxVal) * 5)
		b.WriteString(negativeStyle.Render(fmt.Sprintf(" %s ", strings.Repeat("█", height)+strings.Repeat("░", 5-height))))
	}
	b.WriteString("\n\n")

	// Table view
	b.WriteString("  Month    Revenue    Expenses   Profit\n")
	b.WriteString("  ─────────────────────────────────────\n")
	for i, md := range d.monthlyTrend {
		style := dimStyle
		if i == m.selectedRow {
			style = selectedStyle
		}

		profitStr := fmt.Sprintf("$%.0f", md.profit)
		if md.profit < 0 {
			profitStr = negativeStyle.Render(profitStr)
		} else {
			profitStr = positiveStyle.Render(profitStr)
		}

		line := fmt.Sprintf("  %-6s   $%-8.0f  $%-8.0f  %s", md.month, md.revenue, md.expenses, profitStr)
		b.WriteString(style.Render(line) + "\n")
	}

	return b.String()
}

func (m profitModel) renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return "│" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "│"
}

func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
