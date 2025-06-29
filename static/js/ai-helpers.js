/**
 * Shared JavaScript utilities for AI-generated content (CV and Cover Letter)
 * Used by templates/partials/cv_generator.html and cover_letter_generator.html
 */

window.AIHelpers = (function() {
    'use strict';

    // Common content validation
    function validateContentElement(elementId, contentType = 'content') {
        const element = document.getElementById(elementId);
        if (!element) {
            alert(`Error: ${contentType} not found. Please refresh the page and try again.`);
            return null;
        }
        return element;
    }


    // Copy to clipboard utility
    function copyToClipboard(text) {
        return navigator.clipboard.writeText(text).catch(err => {
            console.error('Failed to copy: ', err);
            alert('Failed to copy to clipboard');
        });
    }

    // Smooth scroll utility
    function scrollToElement(elementId, offset = 0) {
        const element = document.getElementById(elementId);
        if (element) {
            const elementPosition = element.offsetTop + offset;
            window.scrollTo({
                top: elementPosition,
                behavior: 'smooth'
            });
        }
    }

    // Mobile editing helpers
    function enableMobileEditing(containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;

        // Add mobile-friendly classes and behaviors
        container.addEventListener('focus', function(e) {
            if (e.target.contentEditable === 'true') {
                // Scroll element into view on focus for mobile
                if (window.innerWidth < 768) {
                    setTimeout(() => {
                        e.target.scrollIntoView({
                            behavior: 'smooth',
                            block: 'center'
                        });
                    }, 300);
                }
            }
        }, true);
    }

    // Section management (for CV sections)
    function deleteSection(buttonElement) {
        if (!buttonElement) return;

        const section = buttonElement.closest('.resume-section');
        if (section && confirm('Are you sure you want to delete this section?')) {
            section.remove();
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

    // Validation utilities
    function validateRequiredFields(formElement) {
        const requiredFields = formElement.querySelectorAll('[required]');
        let isValid = true;

        requiredFields.forEach(field => {
            if (!field.value.trim()) {
                field.classList.add('border-red-500');
                isValid = false;
            } else {
                field.classList.remove('border-red-500');
            }
        });

        return isValid;
    }

    // Text utilities
    function sanitizeText(text) {
        return text.replace(/\s+/g, ' ').trim();
    }

    function truncateText(text, maxLength) {
        if (text.length <= maxLength) return text;
        return text.substring(0, maxLength) + '...';
    }

    // Custom font loading for jsPDF
    let fontsLoaded = false;

    function loadCustomFonts(doc) {
        if (fontsLoaded) return;

        try {
            // Check if Inter fonts have been loaded via inter-font-loader.js
            if (window.loadInterFonts && typeof window.loadInterFonts === 'function') {
                // Use the external font loader if available
                const success = window.loadInterFonts(doc);
                if (success) {
                    fontsLoaded = true;
                    console.log('Inter fonts loaded successfully from inter-font-loader.js');
                    return;
                }
            }

            // Set Helvetica as the default font (similar to Inter)
            doc.setFont('helvetica', 'normal');
            console.log('Using Helvetica as Inter substitute for PDF generation');

            fontsLoaded = true;

        } catch (e) {
            console.error('Failed to load custom fonts, using defaults:', e);
            doc.setFont('helvetica', 'normal');
        }
    }

    return {
        // PDF utilities
        loadCustomFonts,

        // Content utilities
        validateContentElement,
        copyToClipboard,
        scrollToElement,
        enableMobileEditing,

        // Section management
        deleteSection,

        // Modal utilities
        showModal,
        hideModal,

        // Form utilities
        validateRequiredFields,

        // Text utilities
        sanitizeText,
        truncateText
    };
})();
