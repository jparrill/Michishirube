// Michishirube - Task Detail Page JavaScript
// Handles task-specific interactions on the task detail page

// Tags management
async function addTag(taskId) {
    const input = document.getElementById('new-tag-input');
    const tag = input.value.trim();

    if (!tag) {
        App.notify.warning('Please enter a tag');
        return;
    }

    App.loading.show('Adding tag...');

    try {
        // Get current task data
        const task = await App.api.get(`/api/tasks/${taskId}`);
        const currentTags = task.tags || [];

        if (currentTags.includes(tag)) {
            App.notify.warning('Tag already exists');
            return;
        }

        // Add the new tag
        const updatedTags = [...currentTags, tag];
        await App.api.patch(`/api/tasks/${taskId}`, {
            tags: updatedTags
        });

        // Update UI
        addTagToDOM(tag, taskId);
        input.value = '';

    } catch (error) {
        console.error('Failed to add tag:', error);
        App.notify.error('Failed to add tag');
    } finally {
        App.loading.hide();
    }
}

async function removeTag(taskId, tag) {
    App.loading.show('Removing tag...');

    try {
        // Get current task data
        const task = await App.api.get(`/api/tasks/${taskId}`);
        const currentTags = task.tags || [];

        // Remove the tag
        const updatedTags = currentTags.filter(t => t !== tag);
        await App.api.patch(`/api/tasks/${taskId}`, {
            tags: updatedTags
        });

        // Update UI
        removeTagFromDOM(tag);

    } catch (error) {
        console.error('Failed to remove tag:', error);
        App.notify.error('Failed to remove tag');
    } finally {
        App.loading.hide();
    }
}

function addTagToDOM(tag, taskId) {
    const container = document.querySelector('.tags-container');
    const addForm = container.querySelector('.add-tag-form');

    const tagElement = document.createElement('span');
    tagElement.className = 'tag editable-tag';
    tagElement.innerHTML = `
        ${App.utils.escapeHtml(tag)}
        <button class="tag-remove" onclick="removeTag('${taskId}', '${App.utils.escapeHtml(tag)}')">&times;</button>
    `;

    container.insertBefore(tagElement, addForm);
}

function removeTagFromDOM(tag) {
    const tags = document.querySelectorAll('.tag.editable-tag');
    tags.forEach(tagElement => {
        if (tagElement.textContent.trim().replace('×', '') === tag) {
            tagElement.remove();
        }
    });
}

function handleTagKeypress(event, taskId) {
    if (event.key === 'Enter') {
        event.preventDefault();
        addTag(taskId);
    }
}

// Blockers management
async function addBlocker(taskId) {
    const input = document.getElementById('new-blocker-input');
    const blocker = input.value.trim();

    if (!blocker) {
        App.notify.warning('Please enter a blocker description');
        return;
    }

    App.loading.show('Adding blocker...');

    try {
        // Get current task data
        const task = await App.api.get(`/api/tasks/${taskId}`);
        const currentBlockers = task.blockers || [];

        // Add the new blocker
        const updatedBlockers = [...currentBlockers, blocker];
        await App.api.patch(`/api/tasks/${taskId}`, {
            blockers: updatedBlockers
        });

        // Update UI
        addBlockerToDOM(blocker, taskId, updatedBlockers.length - 1);
        input.value = '';

    } catch (error) {
        console.error('Failed to add blocker:', error);
        App.notify.error('Failed to add blocker');
    } finally {
        App.loading.hide();
    }
}

async function removeBlocker(taskId, index) {
    App.loading.show('Removing blocker...');

    try {
        // Get current task data
        const task = await App.api.get(`/api/tasks/${taskId}`);
        const currentBlockers = task.blockers || [];

        // Remove the blocker
        const updatedBlockers = currentBlockers.filter((_, i) => i !== index);
        await App.api.patch(`/api/tasks/${taskId}`, {
            blockers: updatedBlockers
        });

        // Refresh page to update blocker indices
        window.location.reload();

    } catch (error) {
        console.error('Failed to remove blocker:', error);
        App.notify.error('Failed to remove blocker');
    } finally {
        App.loading.hide();
    }
}

function addBlockerToDOM(blocker, taskId, index) {
    const container = document.querySelector('.blockers-container');
    const addForm = container.querySelector('.add-blocker-form');

    const blockerElement = document.createElement('div');
    blockerElement.className = 'blocker-item';
    blockerElement.innerHTML = `
        <span class="blocker-text">• ${App.utils.escapeHtml(blocker)}</span>
        <button class="blocker-remove" onclick="removeBlocker('${taskId}', ${index})">&times;</button>
    `;

    container.insertBefore(blockerElement, addForm);
}

function handleBlockerKeypress(event, taskId) {
    if (event.key === 'Enter') {
        event.preventDefault();
        addBlocker(taskId);
    }
}

// Links management
function showAddLinkForm(type = 'other') {
    // Hide edit form if it's open
    hideEditLinkForm();
    
    const form = document.getElementById('add-link-form');
    const typeSelect = form.querySelector('#link-type');

    typeSelect.value = type;
    form.style.display = 'block';

    // Focus on URL input
    const urlInput = form.querySelector('#link-url');
    urlInput.focus();
}

function hideAddLinkForm() {
    const form = document.getElementById('add-link-form');
    form.style.display = 'none';

    // Clear form
    form.querySelector('form').reset();
}

function editLink(linkId, type, url, title, status, taskId) {
    // Hide add form if it's open
    hideAddLinkForm();
    
    const form = document.getElementById('edit-link-form');
    
    // Populate form with current values
    document.getElementById('edit-link-id').value = linkId;
    document.getElementById('edit-link-type').value = type;
    document.getElementById('edit-link-url').value = url;
    document.getElementById('edit-link-title').value = title;
    document.getElementById('edit-link-status').value = status;
    
    // Store taskId for later use
    form.dataset.taskId = taskId;
    
    form.style.display = 'block';
    
    // Focus on title input
    const titleInput = form.querySelector('#edit-link-title');
    titleInput.focus();
    titleInput.select();
}

function hideEditLinkForm() {
    const form = document.getElementById('edit-link-form');
    form.style.display = 'none';

    // Clear form
    form.querySelector('form').reset();
}

async function addLink(event, taskId) {
    event.preventDefault();

    const form = event.target;
    const formData = new FormData(form);

    const linkData = {
        task_id: taskId,
        type: formData.get('type'),
        url: formData.get('url'),
        title: formData.get('title') || formData.get('url'),
        status: formData.get('status') || 'active'
    };

    App.loading.show('Adding link...');

    try {
        await App.api.post('/api/links', linkData);

        hideAddLinkForm();

        // Refresh page to show new link
        window.location.reload();

    } catch (error) {
        console.error('Failed to add link:', error);
        App.notify.error('Failed to add link');
    } finally {
        App.loading.hide();
    }
}

async function updateLink(event) {
    event.preventDefault();

    const form = event.target;
    const formData = new FormData(form);
    const editForm = document.getElementById('edit-link-form');

    const linkId = formData.get('id');
    const linkData = {
        id: linkId,
        task_id: editForm.dataset.taskId,  // Get taskId from the form data
        type: formData.get('type'),
        url: formData.get('url'),
        title: formData.get('title') || formData.get('url'),
        status: formData.get('status') || 'active'
    };

    App.loading.show('Updating link...');

    try {
        await App.api.put(`/api/links/${linkId}`, linkData);

        hideEditLinkForm();

        // Refresh page to show updated link
        window.location.reload();

    } catch (error) {
        console.error('Failed to update link:', error);
        App.notify.error('Failed to update link');
    } finally {
        App.loading.hide();
    }
}

async function removeLink(linkId) {
    if (!confirm('Are you sure you want to remove this link?')) {
        return;
    }

    console.log('🔗 Starting link removal for:', linkId);
    App.loading.show('Removing link...');

    try {
        console.log('🔗 Making DELETE request to:', `/api/links/${linkId}`);
        const response = await App.api.delete(`/api/links/${linkId}`);
        console.log('🔗 DELETE request successful, response:', response);

        // Remove the link from UI without reloading the page
        const linkElement = document.querySelector(`[data-link-id="${linkId}"]`);
        console.log('🔗 Found link element:', linkElement);

        if (linkElement) {
            linkElement.remove();
            console.log('🔗 Link element removed from DOM');

            // Check if there are no more links and show "no links" message
            const linksContainer = document.querySelector('.links-container');
            const remainingLinks = linksContainer.querySelectorAll('.link-item');
            const noLinksMessage = linksContainer.querySelector('.no-links');

            console.log('🔗 Remaining links count:', remainingLinks.length);

            if (remainingLinks.length === 0) {
                if (!noLinksMessage) {
                    const noLinks = document.createElement('p');
                    noLinks.className = 'no-links';
                    noLinks.textContent = 'No related links yet.';
                    linksContainer.appendChild(noLinks);
                    console.log('🔗 Added "no links" message');
                }
            }
        } else {
            console.warn('🔗 Link element not found in DOM for ID:', linkId);
        }

        console.log('🔗 Link removal completed successfully');

    } catch (error) {
        console.error('🔗 Failed to remove link:', error);
        console.error('🔗 Error details:', {
            name: error.name,
            message: error.message,
            stack: error.stack
        });
        App.notify.error('Failed to remove link. Please try again.');
    } finally {
        App.loading.hide();
        console.log('🔗 Loading indicator hidden');
    }
}

// Comments management
async function addComment(event, taskId) {
    event.preventDefault();

    const form = event.target;
    const formData = new FormData(form);
    const content = formData.get('content').trim();

    if (!content) {
        App.notify.warning('Please enter a comment');
        return;
    }

    App.loading.show('Adding comment...');

    try {
        await App.api.post('/api/comments', {
            task_id: taskId,
            content: content
        });

        form.reset();

        // Refresh page to show new comment
        window.location.reload();

    } catch (error) {
        console.error('Failed to add comment:', error);
        App.notify.error('Failed to add comment');
    } finally {
        App.loading.hide();
    }
}

async function removeComment(commentId) {
    if (!confirm('Are you sure you want to remove this comment?')) {
        return;
    }

    App.loading.show('Removing comment...');

    try {
        await App.api.delete(`/api/comments/${commentId}`);

        // Remove the comment from UI without reloading the page
        const commentElement = document.querySelector(`[data-comment-id="${commentId}"]`);
        if (commentElement) {
            commentElement.remove();

            // Check if there are no more comments and show "no comments" message
            const commentsContainer = document.querySelector('.comments-container');
            const remainingComments = commentsContainer.querySelectorAll('.comment-item');
            const noCommentsMessage = commentsContainer.querySelector('.no-comments');

            if (remainingComments.length === 0) {
                if (!noCommentsMessage) {
                    const noComments = document.createElement('p');
                    noComments.className = 'no-comments';
                    noComments.textContent = 'No comments yet.';
                    commentsContainer.appendChild(noComments);
                }
            }
        } else {
            console.warn('💬 Comment element not found in DOM for ID:', commentId);
        }

        console.log('💬 Comment removal completed successfully');

    } catch (error) {
        console.error('Failed to remove comment:', error);
        App.notify.error('Failed to remove comment. Please try again.');
    } finally {
        App.loading.hide();
    }
}

// Task actions
async function editTask(taskId) {
    // For now, redirect to a simple form - could be enhanced with modal
    window.location.href = `/edit/${taskId}`;
}

async function deleteTask(taskId) {
    if (!confirm('Are you sure you want to delete this task? This cannot be undone.')) {
        return;
    }

    App.loading.show('Deleting task...');

    try {
        await App.api.delete(`/api/tasks/${taskId}`);

        // Redirect to dashboard
        window.location.href = '/';

    } catch (error) {
        console.error('Failed to delete task:', error);
        App.notify.error('Failed to delete task');
    } finally {
        App.loading.hide();
    }
}

// Initialize task page
document.addEventListener('DOMContentLoaded', () => {
    // Close forms on Escape key
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            hideAddLinkForm();
            hideEditLinkForm();
        }
    });

    // Add event delegation for link buttons
    document.addEventListener('click', (e) => {
        if (e.target.classList.contains('link-edit')) {
            const linkItem = e.target.closest('.link-item');
            if (linkItem) {
                const linkId = linkItem.dataset.linkId;
                const linkType = linkItem.dataset.linkType;
                const linkUrl = linkItem.dataset.linkUrl;
                const linkTitle = linkItem.dataset.linkTitle;
                const linkStatus = linkItem.dataset.linkStatus;
                const taskId = linkItem.dataset.taskId;
                editLink(linkId, linkType, linkUrl, linkTitle, linkStatus, taskId);
            }
        } else if (e.target.classList.contains('link-remove')) {
            const linkItem = e.target.closest('.link-item');
            if (linkItem) {
                const linkId = linkItem.dataset.linkId;
                removeLink(linkId);
            }
        }
    });

    console.log('📋 Task detail page loaded');
});