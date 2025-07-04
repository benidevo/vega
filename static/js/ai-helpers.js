window.AIHelpers = (function() {
    'use strict';

    // Common content validation
    function validateContentElement(elementId, contentType = 'content') {
        const element = document.getElementById(elementId);
        if (!element) {
            if (typeof window.showNotification === 'function') {
                window.showNotification(`${contentType} not found. Please refresh the page and try again.`, 'error', 'Content Error');
            }
            return null;
        }
        return element;
    }


    // Copy to clipboard utility
    function copyToClipboard(text) {
        return navigator.clipboard.writeText(text).catch(err => {
            console.error('Failed to copy: ', err);
            if (typeof window.showNotification === 'function') {
                window.showNotification('Failed to copy to clipboard. Please try again.', 'error', 'Copy Failed');
            }
        });
    }


    // Section management (for CV sections)
    function deleteSection(buttonElement) {
        if (!buttonElement) return;

        const section = buttonElement.closest('.resume-section');
        if (section) {
            const sectionTitle = section.querySelector('h3')?.textContent || 'this section';
            const deleteMessage = document.getElementById('delete-message');
            if (deleteMessage) {
                deleteMessage.textContent = `Are you sure you want to delete "${sectionTitle}"? This action cannot be undone.`;
            }

            // Store the section reference for deletion
            window.itemToDelete = {
                element: section,
                url: null
            };

            showModal('delete-modal-shared');
        }
    }

    // Modal utilities
    function showModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('hidden');
            const firstInput = modal.querySelector('input, textarea, select');
            if (firstInput) {
                setTimeout(() => firstInput.focus(), 100);
            }
        }
    }

    function hideModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('hidden');
        }
    }



    // Add event listener for client-side deletion
    document.addEventListener('client-delete', function(event) {
        if (event.detail && event.detail.element) {
            event.detail.element.remove();
        }
    });

    return {
        // Content utilities
        validateContentElement,
        copyToClipboard,

        // Section management
        deleteSection,

        // Modal utilities
        showModal,
        hideModal
    };
})();
