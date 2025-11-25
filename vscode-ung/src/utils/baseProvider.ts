import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { ErrorHandler } from './errors';
import { NotificationManager } from './notifications';

/**
 * Cache entry with TTL
 */
interface CacheEntry<T> {
    data: T;
    timestamp: number;
}

/**
 * Base tree data provider with common functionality
 */
export abstract class BaseProvider<T extends vscode.TreeItem> implements vscode.TreeDataProvider<T> {
    protected _onDidChangeTreeData: vscode.EventEmitter<T | undefined | null | void> =
        new vscode.EventEmitter<T | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<T | undefined | null | void> =
        this._onDidChangeTreeData.event;

    protected cache: Map<string, CacheEntry<any>> = new Map();
    protected readonly cacheTTL: number;
    protected readonly enableCache: boolean;

    constructor(
        protected cli: UngCli,
        options?: {
            cacheTTL?: number;
            enableCache?: boolean;
        }
    ) {
        this.cacheTTL = options?.cacheTTL ?? 60000; // Default 1 minute
        this.enableCache = options?.enableCache ?? true;
    }

    /**
     * Refresh tree view and clear cache
     */
    refresh(): void {
        this.clearCache();
        this._onDidChangeTreeData.fire();
    }

    /**
     * Get tree item (required by TreeDataProvider)
     */
    getTreeItem(element: T): vscode.TreeItem {
        return element;
    }

    /**
     * Get children (must be implemented by subclasses)
     */
    abstract getChildren(element?: T): Promise<T[]>;

    /**
     * Get cached data or fetch new data
     */
    protected async getCachedData<D>(
        key: string,
        fetchFn: () => Promise<D>,
        ttl?: number
    ): Promise<D> {
        if (!this.enableCache) {
            return fetchFn();
        }

        const cached = this.cache.get(key);
        const now = Date.now();
        const effectiveTTL = ttl ?? this.cacheTTL;

        if (cached && (now - cached.timestamp) < effectiveTTL) {
            ErrorHandler.logInfo(`Cache hit for key: ${key}`);
            return cached.data;
        }

        ErrorHandler.logInfo(`Cache miss for key: ${key}, fetching data...`);
        const data = await fetchFn();
        this.cache.set(key, { data, timestamp: now });
        return data;
    }

    /**
     * Clear all cache
     */
    protected clearCache(): void {
        this.cache.clear();
        ErrorHandler.logInfo(`Cache cleared for ${this.constructor.name}`);
    }

    /**
     * Clear specific cache entry
     */
    protected clearCacheEntry(key: string): void {
        this.cache.delete(key);
        ErrorHandler.logInfo(`Cache entry cleared: ${key}`);
    }

    /**
     * Handle errors consistently
     */
    protected handleError(error: any, context: string): T[] {
        ErrorHandler.handle(error, context);
        return this.getErrorItem(error) as T[];
    }

    /**
     * Get error tree item
     */
    protected getErrorItem(error: any): vscode.TreeItem {
        const item = new vscode.TreeItem(
            'Error loading data',
            vscode.TreeItemCollapsibleState.None
        );
        item.description = 'Click to retry';
        item.tooltip = error.message || String(error);
        item.iconPath = new vscode.ThemeIcon('error');
        item.command = {
            command: 'workbench.action.reloadWindow',
            title: 'Retry'
        };
        return item;
    }

    /**
     * Get loading tree item
     */
    protected getLoadingItem(): vscode.TreeItem {
        const item = new vscode.TreeItem(
            'Loading...',
            vscode.TreeItemCollapsibleState.None
        );
        item.iconPath = new vscode.ThemeIcon('loading~spin');
        return item;
    }

    /**
     * Get empty state tree item
     */
    protected getEmptyItem(message: string, actionCommand?: string): vscode.TreeItem {
        const item = new vscode.TreeItem(
            message,
            vscode.TreeItemCollapsibleState.None
        );
        item.iconPath = new vscode.ThemeIcon('inbox');
        item.contextValue = 'empty';

        if (actionCommand) {
            item.command = {
                command: actionCommand,
                title: 'Create'
            };
        }

        return item;
    }

    /**
     * Execute with loading state
     */
    protected async executeWithLoading<R>(
        operation: () => Promise<R>,
        loadingMessage?: string
    ): Promise<R> {
        return NotificationManager.withProgress(
            loadingMessage || 'Loading...',
            operation
        );
    }

    /**
     * Filter items based on search query
     */
    protected filterItems(items: T[], searchQuery?: string): T[] {
        if (!searchQuery || searchQuery.trim() === '') {
            return items;
        }

        const query = searchQuery.toLowerCase();
        return items.filter(item => {
            const label = item.label?.toString().toLowerCase() || '';
            const description = item.description?.toString().toLowerCase() || '';
            return label.includes(query) || description.includes(query);
        });
    }

    /**
     * Sort items by field
     */
    protected sortItems(
        items: T[],
        field: keyof T,
        direction: 'asc' | 'desc' = 'asc'
    ): T[] {
        return items.sort((a, b) => {
            const aVal = a[field];
            const bVal = b[field];

            if (aVal === undefined || aVal === null || bVal === undefined || bVal === null) {
                return 0;
            }

            let comparison = 0;
            if (aVal < bVal) {
                comparison = -1;
            } else if (aVal > bVal) {
                comparison = 1;
            }

            return direction === 'asc' ? comparison : -comparison;
        });
    }

    /**
     * Group items by a field
     */
    protected groupItems<K extends string | number>(
        items: T[],
        getGroupKey: (item: T) => K
    ): Map<K, T[]> {
        const groups = new Map<K, T[]>();

        for (const item of items) {
            const key = getGroupKey(item);
            const group = groups.get(key) || [];
            group.push(item);
            groups.set(key, group);
        }

        return groups;
    }

    /**
     * Dispose provider and clear cache
     */
    dispose(): void {
        this.clearCache();
        this._onDidChangeTreeData.dispose();
    }
}

/**
 * Base class for grouped providers (with sections)
 */
export abstract class GroupedProvider<
    TSection extends vscode.TreeItem,
    TItem extends vscode.TreeItem
> extends BaseProvider<TSection | TItem> {

    /**
     * Get root sections
     */
    protected abstract getSections(): Promise<TSection[]>;

    /**
     * Get items for a section
     */
    protected abstract getSectionItems(section: TSection): Promise<TItem[]>;

    /**
     * Get children - handles both sections and items
     */
    async getChildren(element?: TSection | TItem): Promise<(TSection | TItem)[]> {
        if (!element) {
            // Root level - return sections
            try {
                return await this.getSections();
            } catch (error) {
                return this.handleError(error, 'Failed to load sections');
            }
        }

        // Check if element is a section
        if (this.isSection(element)) {
            try {
                return await this.getSectionItems(element as TSection);
            } catch (error) {
                return this.handleError(error, 'Failed to load items');
            }
        }

        // Leaf item - no children
        return [];
    }

    /**
     * Check if element is a section (must be implemented by subclass)
     */
    protected abstract isSection(element: TSection | TItem): element is TSection;
}
