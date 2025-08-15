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
        App.notify.success('Status updated successfully');
        
        // Optionally refresh the page or update UI
        setTimeout(() => window.location.reload(), 1000);
        
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
        App.notify.success(`${field.charAt(0).toUpperCase() + field.slice(1)} updated successfully`);
        
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

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    initializeSearch();
    initializeKeyboardShortcuts();
    
    console.log('🗯️ Michishirube loaded');
});

// Export for use in other modules
window.App = App;