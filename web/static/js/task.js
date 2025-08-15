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
        App.notify.success('Tag added successfully');
        
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
        App.notify.success('Tag removed successfully');
        
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
        App.notify.success('Blocker added successfully');
        
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
        App.notify.success('Link added successfully');
        
        // Refresh page to show new link
        setTimeout(() => window.location.reload(), 1000);
        
    } catch (error) {
        console.error('Failed to add link:', error);
        App.notify.error('Failed to add link');
    } finally {
        App.loading.hide();
    }
}

async function removeLink(linkId) {
    if (!confirm('Are you sure you want to remove this link?')) {
        return;
    }

    App.loading.show('Removing link...');
    
    try {
        await App.api.delete(`/api/links/${linkId}`);
        
        App.notify.success('Link removed successfully');
        
        // Refresh page to update UI
        setTimeout(() => window.location.reload(), 1000);
        
    } catch (error) {
        console.error('Failed to remove link:', error);
        App.notify.error('Failed to remove link');
    } finally {
        App.loading.hide();
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
        App.notify.success('Comment added successfully');
        
        // Refresh page to show new comment
        setTimeout(() => window.location.reload(), 1000);
        
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
        
        App.notify.success('Comment removed successfully');
        
        // Refresh page to update UI
        setTimeout(() => window.location.reload(), 1000);
        
    } catch (error) {
        console.error('Failed to remove comment:', error);
        App.notify.error('Failed to remove comment');
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
        
        App.notify.success('Task deleted successfully');
        
        // Redirect to dashboard
        setTimeout(() => window.location.href = '/', 1000);
        
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
        }
    });
    
    console.log('📋 Task detail page loaded');
});