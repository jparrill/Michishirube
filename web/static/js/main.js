// Michishirube - Main JavaScript
// Provides core functionality for all pages

// Global state and utilities
const App = {
    // API helper
    api: {
        async request(url, options = {}) {
            const defaultOptions = {
                headers: {
                    'Content-Type': 'application/json',
                },
            };
            
            const config = { ...defaultOptions, ...options };
            
            try {
                const response = await fetch(url, config);
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return await response.json();
            } catch (error) {
                console.error('API request failed:', error);
                throw error;
            }
        },

        async get(url) {
            return this.request(url);
        },

        async post(url, data) {
            return this.request(url, {
                method: 'POST',
                body: JSON.stringify(data),
            });
        },

        async patch(url, data) {
            return this.request(url, {
                method: 'PATCH',
                body: JSON.stringify(data),
            });
        },

        async delete(url) {
            return this.request(url, {
                method: 'DELETE',
            });
        }
    },

    // Loading indicator
    loading: {
        show(text = 'Loading...') {
            const indicator = document.getElementById('loading-indicator') || 
                            document.getElementById('status-loading');
            if (indicator) {
                const textElement = indicator.querySelector('#loading-text') || 
                                  indicator.querySelector('span');
                if (textElement) {
                    textElement.textContent = text;
                }
                indicator.style.display = 'flex';
            }
        },

        hide() {
            const indicators = [
                document.getElementById('loading-indicator'),
                document.getElementById('status-loading')
            ];
            indicators.forEach(indicator => {
                if (indicator) {
                    indicator.style.display = 'none';
                }
            });
        }
    },

    // Notification system
    notify: {
        show(message, type = 'info') {
            // Simple alert for now - can be enhanced later
            const emoji = {
                success: '✅',
                error: '❌',
                warning: '⚠️',
                info: 'ℹ️'
            };
            
            alert(`${emoji[type] || emoji.info} ${message}`);
        },

        success(message) {
            this.show(message, 'success');
        },

        error(message) {
            this.show(message, 'error');
        },

        warning(message) {
            this.show(message, 'warning');
        }
    },

    // Utility functions
    utils: {
        debounce(func, wait) {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    clearTimeout(timeout);
                    func(...args);
                };
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
            };
        },

        formatDate(date) {
            return new Date(date).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        },

        escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    }
};

// Global functions used by templates

// Status update function for dashboard
async function updateTaskStatus(selectElement) {
    const taskId = selectElement.dataset.taskId;
    const newStatus = selectElement.value;
    const originalValue = selectElement.defaultValue;

    App.loading.show('Updating status...');

    try {
        await App.api.patch(`/api/tasks/${taskId}`, {
            status: newStatus
        });

        selectElement.defaultValue = newStatus;
        
        // Refresh the page to update UI
        window.location.reload();
        
    } catch (error) {
        console.error('Failed to update status:', error);
        selectElement.value = originalValue; // Revert change
        App.notify.error('Failed to update status');
    } finally {
        App.loading.hide();
    }
}

// Generic task field update for task detail page
async function updateTaskField(selectElement, field) {
    const taskId = selectElement.dataset.taskId;
    const newValue = selectElement.value;
    const originalValue = selectElement.defaultValue;

    App.loading.show('Updating...');

    try {
        await App.api.patch(`/api/tasks/${taskId}`, {
            [field]: newValue
        });

        selectElement.defaultValue = newValue;
        
    } catch (error) {
        console.error(`Failed to update ${field}:`, error);
        selectElement.value = originalValue; // Revert change
        App.notify.error(`Failed to update ${field}`);
    } finally {
        App.loading.hide();
    }
}

// Search functionality
function initializeSearch() {
    const searchInput = document.getElementById('search-input');
    if (searchInput) {
        // Auto-submit search after user stops typing
        const debouncedSearch = App.utils.debounce(() => {
            if (searchInput.value.length > 2 || searchInput.value.length === 0) {
                searchInput.form.submit();
            }
        }, 500);

        searchInput.addEventListener('input', debouncedSearch);
    }
}

// Keyboard shortcuts
function initializeKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
        // Ignore if user is typing in an input
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
            return;
        }

        switch (e.key) {
            case 'n':
                if (e.ctrlKey || e.metaKey) {
                    e.preventDefault();
                    window.location.href = '/new';
                }
                break;
            case '/':
                e.preventDefault();
                const searchInput = document.getElementById('search-input');
                if (searchInput) {
                    searchInput.focus();
                }
                break;
            case 'Escape':
                const forms = document.querySelectorAll('[style*="display: none"]');
                forms.forEach(form => {
                    if (form.id === 'add-link-form' || form.id === 'validation-messages') {
                        form.style.display = 'none';
                    }
                });
                break;
        }
    });
}

// Report functionality
async function generateReport() {
    App.loading.show('Generating report...');
    
    try {
        const reportData = await App.api.get('/api/report');
        
        let reportText = '## 📊 Weekly Status Report\n\n';
        
        // Working on section
        reportText += '### 🦀 Things I\'ve been working on\n';
        if (reportData.working_on && reportData.working_on.length > 0) {
            reportData.working_on.forEach(task => {
                const jiraId = task.jira_id !== 'NO-JIRA' ? `[${task.jira_id}] ` : '';
                reportText += `- ${jiraId}${task.title}\n`;
                
                // Add links if they exist, grouped by type
                if (task.links && task.links.length > 0) {
                    const linkGroups = {};
                    
                    // Group links by type
                    task.links.forEach(link => {
                        if (!linkGroups[link.type]) {
                            linkGroups[link.type] = [];
                        }
                        linkGroups[link.type].push(link);
                    });
                    
                    // Format each group
                    Object.keys(linkGroups).forEach(type => {
                        const links = linkGroups[type];
                        let prefix = '';
                        let formattedLinks = [];
                        
                        if (type === 'pull_request') {
                            prefix = 'PRs: ';
                            formattedLinks = links.map(link => {
                                const urlPath = link.url.split('/').pop();
                                return `[${urlPath}](${link.url})`;
                            });
                        } else if (type === 'slack_thread') {
                            prefix = 'Slack: ';
                            formattedLinks = links.map((link, index) => {
                                return `[thread-${index + 1}](${link.url})`;
                            });
                        } else if (type === 'jira_ticket') {
                            prefix = 'Jira related: ';
                            formattedLinks = links.map(link => {
                                const urlPath = link.url.split('/').pop();
                                return `[${urlPath}](${link.url})`;
                            });
                        } else {
                            prefix = `${type}: `;
                            formattedLinks = links.map(link => {
                                // For others, use the URL filename without extension as link text
                                const urlPath = link.url.split('/').pop();
                                const linkText = urlPath.includes('.') ? urlPath.split('.')[0] : urlPath;
                                return `[${linkText}](${link.url})`;
                            });
                        }
                        
                        reportText += `  - ${prefix}${formattedLinks.join(', ')}\n`;
                    });
                }
            });
        } else {
            reportText += '- No current work items\n';
        }
        
        // Next up section
        reportText += '\n### 🖖 Things I plan on working on next\n';
        if (reportData.next_up && reportData.next_up.length > 0) {
            reportData.next_up.forEach(task => {
                const jiraId = task.jira_id !== 'NO-JIRA' ? `[${task.jira_id}] ` : '';
                reportText += `- ${jiraId}${task.title}\n`;
                
                // Add links if they exist, grouped by type
                if (task.links && task.links.length > 0) {
                    const linkGroups = {};
                    
                    // Group links by type
                    task.links.forEach(link => {
                        if (!linkGroups[link.type]) {
                            linkGroups[link.type] = [];
                        }
                        linkGroups[link.type].push(link);
                    });
                    
                    // Format each group
                    Object.keys(linkGroups).forEach(type => {
                        const links = linkGroups[type];
                        let prefix = '';
                        let formattedLinks = [];
                        
                        if (type === 'pull_request') {
                            prefix = 'PRs: ';
                            formattedLinks = links.map(link => {
                                const urlPath = link.url.split('/').pop();
                                return `[${urlPath}](${link.url})`;
                            });
                        } else if (type === 'slack_thread') {
                            prefix = 'Slack: ';
                            formattedLinks = links.map((link, index) => {
                                return `[thread-${index + 1}](${link.url})`;
                            });
                        } else if (type === 'jira_ticket') {
                            prefix = 'Jira related: ';
                            formattedLinks = links.map(link => {
                                const urlPath = link.url.split('/').pop();
                                return `[${urlPath}](${link.url})`;
                            });
                        } else {
                            prefix = `${type}: `;
                            formattedLinks = links.map(link => {
                                // For others, use the URL filename without extension as link text
                                const urlPath = link.url.split('/').pop();
                                const linkText = urlPath.includes('.') ? urlPath.split('.')[0] : urlPath;
                                return `[${linkText}](${link.url})`;
                            });
                        }
                        
                        reportText += `  - ${prefix}${formattedLinks.join(', ')}\n`;
                    });
                }
            });
        } else {
            reportText += '- No high priority items planned\n';
        }
        
        // Blockers section
        reportText += '\n### 🤦 Things that are blocking me\n';
        if (reportData.blockers && reportData.blockers.length > 0) {
            reportData.blockers.forEach(task => {
                const jiraId = task.jira_id !== 'NO-JIRA' ? `[${task.jira_id}] ` : '';
                reportText += `- ${jiraId}${task.title}\n`;
                
                // Add blockers if they exist
                if (task.blockers && task.blockers.length > 0) {
                    task.blockers.forEach(blocker => {
                        reportText += `  - ⚠️ ${blocker}\n`;
                    });
                }
                
                // Add links if they exist, grouped by type
                if (task.links && task.links.length > 0) {
                    const linkGroups = {};
                    
                    // Group links by type
                    task.links.forEach(link => {
                        if (!linkGroups[link.type]) {
                            linkGroups[link.type] = [];
                        }
                        linkGroups[link.type].push(link);
                    });
                    
                    // Format each group
                    Object.keys(linkGroups).forEach(type => {
                        const links = linkGroups[type];
                        let prefix = '';
                        let formattedLinks = [];
                        
                        if (type === 'pull_request') {
                            prefix = 'PRs: ';
                            formattedLinks = links.map(link => {
                                const urlPath = link.url.split('/').pop();
                                return `[${urlPath}](${link.url})`;
                            });
                        } else if (type === 'slack_thread') {
                            prefix = 'Slack: ';
                            formattedLinks = links.map((link, index) => {
                                return `[thread-${index + 1}](${link.url})`;
                            });
                        } else if (type === 'jira_ticket') {
                            prefix = 'Jira related: ';
                            formattedLinks = links.map(link => {
                                const urlPath = link.url.split('/').pop();
                                return `[${urlPath}](${link.url})`;
                            });
                        } else {
                            prefix = `${type}: `;
                            formattedLinks = links.map(link => {
                                // For others, use the URL filename without extension as link text
                                const urlPath = link.url.split('/').pop();
                                const linkText = urlPath.includes('.') ? urlPath.split('.')[0] : urlPath;
                                return `[${linkText}](${link.url})`;
                            });
                        }
                        
                        reportText += `  - ${prefix}${formattedLinks.join(', ')}\n`;
                    });
                }
            });
        } else {
            reportText += '- No current blockers\n';
        }
        
        // Copy to clipboard
        await navigator.clipboard.writeText(reportText);
        
    } catch (error) {
        console.error('Failed to generate report:', error);
        App.notify.error('Failed to generate report');
    } finally {
        App.loading.hide();
    }
}

// Dropdown functionality
function toggleReportDropdown() {
    const reportDropdown = document.getElementById('report-dropdown');
    const filtersDropdown = document.getElementById('filters-dropdown');
    
    // Close filters dropdown if open
    filtersDropdown.classList.remove('show');
    
    // Toggle report dropdown
    reportDropdown.classList.toggle('show');
}

function toggleFiltersDropdown() {
    const reportDropdown = document.getElementById('report-dropdown');
    const filtersDropdown = document.getElementById('filters-dropdown');
    
    // Close report dropdown if open
    reportDropdown.classList.remove('show');
    
    // Toggle filters dropdown
    filtersDropdown.classList.toggle('show');
}

// Indicator dropdown functionality
function toggleIndicatorDropdown(element) {
    const dropdown = element.parentElement;
    const menu = dropdown.querySelector('.indicator-dropdown-menu');
    
    // Close all other indicator dropdowns
    document.querySelectorAll('.indicator-dropdown-menu.show').forEach(otherMenu => {
        if (otherMenu !== menu) {
            otherMenu.classList.remove('show');
        }
    });
    
    // Toggle current dropdown
    menu.classList.toggle('show');
    
    // Prevent event bubbling
    event.stopPropagation();
}

// Close dropdown when clicking outside
document.addEventListener('click', (event) => {
    const reportDropdown = document.getElementById('report-dropdown');
    const filtersDropdown = document.getElementById('filters-dropdown');
    const button = event.target.closest('.dropdown-toggle');
    const indicatorButton = event.target.closest('.indicator-dropdown-toggle');
    
    if (!button) {
        if (reportDropdown) reportDropdown.classList.remove('show');
        if (filtersDropdown) filtersDropdown.classList.remove('show');
    }
    
    // Close indicator dropdowns when clicking outside
    if (!indicatorButton && !event.target.closest('.indicator-dropdown-menu')) {
        document.querySelectorAll('.indicator-dropdown-menu.show').forEach(menu => {
            menu.classList.remove('show');
        });
    }
});

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    initializeSearch();
    initializeKeyboardShortcuts();
    
    console.log('🗯️ Michishirube loaded');
});

// Export for use in other modules
window.App = App;