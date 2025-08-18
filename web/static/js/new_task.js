// Michishirube - New Task Form JavaScript
// Handles the new task creation form functionality

// Global state for the form
let linkCounter = 0;

// Link management functions
function addLinkRow(type = 'other') {
    const container = document.getElementById('links-container');
    const linkRow = document.createElement('div');
    linkRow.className = 'link-input-row';
    linkRow.innerHTML = `
        <select name="link_types[]">
            <option value="pull_request" ${type === 'pull_request' ? 'selected' : ''}>🔗 Pull Request</option>
            <option value="slack_thread" ${type === 'slack_thread' ? 'selected' : ''}>💬 Slack Thread</option>
            <option value="jira_ticket" ${type === 'jira_ticket' ? 'selected' : ''}>📋 Jira Ticket</option>
            <option value="documentation" ${type === 'documentation' ? 'selected' : ''}>📚 Documentation</option>
            <option value="other" ${type === 'other' ? 'selected' : ''}>🌐 Other</option>
        </select>
        <input type="url" name="link_urls[]" placeholder="https://..." required>
        <input type="text" name="link_titles[]" placeholder="Title (optional)">
        <button type="button" class="btn btn-danger btn-small" onclick="removeLinkRow(this)">Remove</button>
    `;

    container.appendChild(linkRow);
    linkCounter++;
}

function removeLinkRow(button) {
    const linkRow = button.closest('.link-input-row');
    if (linkRow) {
        linkRow.remove();
    }
}

// Tag management functions
function addFormTag() {
    const input = document.getElementById('tag-input');
    const tag = input.value.trim();

    if (tag && !isTagAlreadyAdded(tag)) {
        addTagToDisplay(tag);
        updateHiddenTags();
        input.value = '';
    }
}

function removeFormTag(tag) {
    const tagElement = document.querySelector(`.tag:has(.tag-remove[onclick*="${tag}"])`);
    if (tagElement) {
        tagElement.remove();
        updateHiddenTags();
    }
}

function isTagAlreadyAdded(tag) {
    const existingTags = document.querySelectorAll('.tag');
    for (let tagElement of existingTags) {
        if (tagElement.textContent.trim().replace('×', '').trim() === tag) {
            return true;
        }
    }
    return false;
}

function addTagToDisplay(tag) {
    const tagsDisplay = document.getElementById('tags-display');
    const tagSpan = document.createElement('span');
    tagSpan.className = 'tag';
    tagSpan.innerHTML = `
        ${tag}
        <button type="button" class="tag-remove" onclick="removeFormTag('${tag}')">&times;</button>
    `;
    tagsDisplay.appendChild(tagSpan);
}

function updateHiddenTags() {
    const tags = Array.from(document.querySelectorAll('.tag'))
        .map(tag => tag.textContent.trim().replace('×', '').trim())
        .filter(tag => tag.length > 0);

    const hiddenInput = document.getElementById('tags-hidden');
    hiddenInput.value = tags.join(',');
}

function handleFormTagKeypress(event) {
    if (event.key === 'Enter') {
        event.preventDefault();
        addFormTag();
    }
}

// Form validation and submission
function validateForm() {
    const errors = [];

    // Check required fields
    const title = document.getElementById('title').value.trim();
    if (!title) {
        errors.push('Task title is required');
    }

    // Check link URLs if any are provided
    const linkUrls = document.querySelectorAll('input[name="link_urls[]"]');
    for (let urlInput of linkUrls) {
        if (urlInput.value.trim() && !isValidURL(urlInput.value.trim())) {
            errors.push('Please enter valid URLs for all links');
            break;
        }
    }

    return errors;
}

function isValidURL(string) {
    try {
        new URL(string);
        return true;
    } catch (_) {
        return false;
    }
}

function showValidationMessages(errors) {
    const container = document.getElementById('validation-messages');
    const list = document.getElementById('validation-list');

    list.innerHTML = '';
    errors.forEach(error => {
        const li = document.createElement('li');
        li.textContent = error;
        list.appendChild(li);
    });

    container.style.display = 'block';
}

function hideValidationMessages() {
    const container = document.getElementById('validation-messages');
    container.style.display = 'none';
}

// Form submission handling
function handleFormSubmit(event) {
    const errors = validateForm();

    if (errors.length > 0) {
        event.preventDefault();
        showValidationMessages(errors);
        return false;
    }

    // Hide any existing validation messages
    hideValidationMessages();
    return true;
}

// Draft functionality
function saveDraft() {
    // For now, just show a message that this feature is coming soon
    alert('💾 Draft functionality coming soon! For now, please complete and submit the form.');
}

// Initialize form when page loads
document.addEventListener('DOMContentLoaded', function() {
    const form = document.querySelector('form');
    if (form) {
        form.addEventListener('submit', handleFormSubmit);
    }

    // Initialize tags display if any exist
    updateHiddenTags();

    // Add event listener for tag input
    const tagInput = document.getElementById('tag-input');
    if (tagInput) {
        tagInput.addEventListener('blur', addFormTag);
    }
});

// Export functions for global access
window.addLinkRow = addLinkRow;
window.removeLinkRow = removeLinkRow;
window.addFormTag = addFormTag;
window.removeFormTag = removeFormTag;
window.handleFormTagKeypress = handleFormTagKeypress;
window.saveDraft = saveDraft;
window.hideValidationMessages = hideValidationMessages;
