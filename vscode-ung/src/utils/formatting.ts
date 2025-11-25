import { Config } from './config';

/**
 * Formatting utilities
 */
export class Formatter {
    /**
     * Format a date according to user preferences
     */
    static formatDate(date: Date | string): string {
        const dateObj = typeof date === 'string' ? new Date(date) : date;
        const format = Config.getDateFormat();

        const year = dateObj.getFullYear();
        const month = String(dateObj.getMonth() + 1).padStart(2, '0');
        const day = String(dateObj.getDate()).padStart(2, '0');

        switch (format) {
            case 'MM/DD/YYYY':
                return `${month}/${day}/${year}`;
            case 'DD/MM/YYYY':
                return `${day}/${month}/${year}`;
            case 'YYYY-MM-DD':
            default:
                return `${year}-${month}-${day}`;
        }
    }

    /**
     * Format a currency amount
     */
    static formatCurrency(amount: number, currency: string = 'USD'): string {
        const currencySymbols: Record<string, string> = {
            'USD': '$',
            'EUR': '€',
            'GBP': '£',
            'UAH': '₴',
            'CAD': 'C$',
            'AUD': 'A$'
        };

        const symbol = currencySymbols[currency] || currency;
        return `${symbol}${amount.toFixed(2)}`;
    }

    /**
     * Format duration from seconds to human-readable format
     */
    static formatDuration(seconds: number): string {
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;

        if (hours > 0) {
            return `${hours}h ${minutes}m`;
        } else if (minutes > 0) {
            return `${minutes}m ${secs}s`;
        } else {
            return `${secs}s`;
        }
    }

    /**
     * Format hours to human-readable format
     */
    static formatHours(hours: number): string {
        const h = Math.floor(hours);
        const m = Math.round((hours - h) * 60);

        if (h > 0 && m > 0) {
            return `${h}h ${m}m`;
        } else if (h > 0) {
            return `${h}h`;
        } else {
            return `${m}m`;
        }
    }

    /**
     * Parse date string to ISO format
     */
    static parseDate(dateStr: string): string {
        const date = new Date(dateStr);
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        return `${year}-${month}-${day}`;
    }

    /**
     * Get status badge emoji
     */
    static getStatusBadge(status: string): string {
        const badges: Record<string, string> = {
            'pending': '⏳',
            'paid': '✅',
            'overdue': '⚠️',
            'draft': '📝',
            'cancelled': '❌'
        };
        return badges[status.toLowerCase()] || '•';
    }

    /**
     * Truncate text to a maximum length
     */
    static truncate(text: string, maxLength: number): string {
        if (text.length <= maxLength) {
            return text;
        }
        return text.substring(0, maxLength - 3) + '...';
    }
}
